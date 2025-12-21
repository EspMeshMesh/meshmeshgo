package meshmesh

import (
	"fmt"
	"net"

	"github.com/charmbracelet/log"
	"golang.org/x/exp/slices"
	"leguru.net/m/v2/graph"
	"leguru.net/m/v2/logger"
	"leguru.net/m/v2/utils"
)

var _allStats *EspApiStats

/*const (
	esphomeapiWaitPacketHead int = 0
	esphomeapiWaitPacketSize int = 1
	esphomeapiWaitPacketData int = 3
)*/

const (
	fixedApiRemotePort int = 6053
	fixedOtaRemotePort int = 3232
)

type ServerApi struct {
	Address       MeshNodeId
	Clients       []*ConnectionPathBridge
	listener      net.Listener
	listenAddress string
	network       *graph.Network
}

func (s *ServerApi) connectionFactory(remotePort int) ConnectionPathBridgeDriver {
	var client ConnectionPathBridgeDriver = nil

	if remotePort == fixedApiRemotePort {
		client = NewApiConnection()
	} else {
		client = NewOtaConnection()
	}

	return client
}

func (s *ServerApi) GetListenAddress() string {
	return s.listenAddress
}

func (s *ServerApi) ClientClosedCb(client *ConnectionPathBridge) {
	// Remove client from clients list
	client.Dispose()
	idx := slices.Index(s.Clients, client)
	if idx >= 0 {
		s.Clients = append(s.Clients[:idx], s.Clients[idx+1:]...)
	}
	logger.Info("Closed EspHomeApi connection")
}

func (s *ServerApi) ListenAndServe(serialProxy *ConnectedPath2Serial, remotePort int) {
	for {
		socket, err := s.listener.Accept()
		if err != nil {
			logger.Error(err)
			continue
		}

		logger.WithFields(logger.Fields{"nodeId": s.Address, "active": len(s.Clients)}).Debug("EspHome connection accepted")

		driver := s.connectionFactory(remotePort)
		client, err := NewConnectionPathBridge(socket, serialProxy, s.network, s.Address, remotePort, driver, s.ClientClosedCb)
		if err == nil {
			s.Clients = append(s.Clients, client)
			logger.WithFields(logger.Fields{"nodeId": utils.FmtNodeId(int64(s.Address)), "clients": len(s.Clients)}).Debug("Added new client")
		} else {
			logger.Error(err)
			socket.Close()
		}
	}
}

func (s *ServerApi) CloseConnections() {
	for _, client := range s.Clients {
		client.close()
	}
}

func (s *ServerApi) ShutDown() {
	s.listener.Close()
}

func NewServerSocket(serialProxy *ConnectedPath2Serial, network *graph.Network, address MeshNodeId, config *ServerApiConfig) (*ServerApi, error) {
	var bindAddress string = config.BindAddress
	if config.BindAddress == "" || config.BindAddress == "dynamic" {
		bindAddress = utils.FmtNodeIdHass(int64(address))
	}

	bindPort := config.BindPort
	if config.BindPort <= 0 {
		bindPort = utils.HashString(utils.FmtNodeId(int64(address)), config.SizeOfPortsPool) + config.BasePortOffset
	}

	server := ServerApi{Address: address, network: network}
	server.listenAddress = fmt.Sprintf("%s:%d", bindAddress, bindPort)
	listener, err := net.Listen("tcp4", server.listenAddress)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	logger.WithFields(logger.Fields{"node": utils.FmtNodeId(int64(address)), "bind": server.listenAddress}).Debug("Start listening on port for node connection")
	server.listener = listener
	go server.ListenAndServe(serialProxy, config.RemotePort)
	return &server, nil
}

func (m *MultiServerApi) Stats() *EspApiStats {
	return m.espApiStats
}

func (m *MultiServerApi) PrintStats() {
	m.espApiStats.PrintStats()
}

func (m *MultiServerApi) CloseConnection(addr MeshNodeId) {
	for _, server := range m.Servers {
		if server.Address == addr {
			server.CloseConnections()
		}
	}
}

func (m *MultiServerApi) MainNetworkChanged(network *graph.Network) {
	nodes := network.Nodes()
	for nodes.Next() {
		node := nodes.Node().(graph.NodeDevice)
		if node.Device().InUse() {
			found := false
			for _, server := range m.Servers {
				if server.Address == MeshNodeId(node.ID()) {
					found = true
					break
				}
			}
			if !found {
				logger.WithFields(logger.Fields{"node": utils.FmtNodeId(int64(node.ID()))}).Debug("MainNetworkChanged adding esphome connection to new node")
				server, err := NewServerSocket(m.serialProxy, network, MeshNodeId(node.ID()), &m.config)
				if err != nil {
					log.Error(err)
				} else {
					m.Servers = append(m.Servers, server)
				}
			}
		}
	}

	newServers := make([]*ServerApi, 0)
	for _, server := range m.Servers {
		found := false
		nodes = graph.GetMainNetwork().Nodes()
		for nodes.Next() {
			node := nodes.Node().(graph.NodeDevice)
			if server.Address == MeshNodeId(node.ID()) {
				found = true
				break
			}
		}
		if !found {
			logger.WithFields(logger.Fields{"server": server.Address}).Debug("MainNetworkChanged deleting esphome connection to non existing node")
			server.CloseConnections()
		} else {
			newServers = append(newServers, server)
		}
	}
	m.Servers = newServers
}

func (m *MultiServerApi) StarPathProtocol(starpath *StarPath) {
	m.starPath = starpath
	starpath.network.AddNetworkChangedCallback(m.starPathNetworkChanged)
	m.starPathNetworkChanged(starpath.network)
}

func (m *MultiServerApi) serverAddressExists(nodeId MeshNodeId) bool {
	for _, server := range m.Servers {
		if server.Address == nodeId {
			return true
		}
	}
	return false
}

func (m *MultiServerApi) getNewNodes(network *graph.Network) []MeshNodeId {
	newNodes := make([]MeshNodeId, 0)
	nodes := network.Nodes()
	for nodes.Next() {
		node := nodes.Node().(graph.NodeDevice)
		if node.Device().InUse() && !network.IsLocalDevice(node) {
			if !m.serverAddressExists(MeshNodeId(node.ID())) {
				newNodes = append(newNodes, MeshNodeId(node.ID()))
			}
		}
	}
	return newNodes
}

func (m *MultiServerApi) createApiAndOtaServers(nodeId MeshNodeId, network *graph.Network) {
	configApi := m.config
	configApi.RemotePort = fixedApiRemotePort

	server, err := NewServerSocket(m.serialProxy, network, nodeId, &configApi)
	if err != nil {
		log.Error(err)
	} else {
		m.Servers = append(m.Servers, server)
	}

	configOta := m.config
	configOta.BindPort = fixedOtaRemotePort
	configOta.RemotePort = fixedOtaRemotePort
	serverOta, err := NewServerSocket(m.serialProxy, network, nodeId, &configOta)
	if err != nil {
		log.Error(err)
	} else {
		m.Servers = append(m.Servers, serverOta)
	}
}

func (m *MultiServerApi) starPathNetworkChanged(network *graph.Network) {
	newNodes := m.getNewNodes(network)
	for _, nodeId := range newNodes {
		logger.WithFields(logger.Fields{"nodeId": nodeId}).Debug("starPathNetworkChanged adding new node to star path")
		m.createApiAndOtaServers(nodeId, network)
	}
}

type ServerApiConfig struct {
	BindAddress     string
	BindPort        int
	RemotePort      int
	BasePortOffset  int
	SizeOfPortsPool int
}

type MultiServerApi struct {
	espApiStats *EspApiStats
	serialProxy *ConnectedPath2Serial
	starPath    *StarPath
	config      ServerApiConfig
	Servers     []*ServerApi
}

func NewMultiSocketServer(serialProxy *ConnectedPath2Serial, config ServerApiConfig) *MultiServerApi {
	_allStats = NewEspApiStats()
	multisrv := MultiServerApi{serialProxy: serialProxy, espApiStats: _allStats, config: config}
	serialProxy.ClearConnections()

	network := graph.GetMainNetwork()
	nodes := network.Nodes()
	network.AddNetworkChangedCallback(multisrv.MainNetworkChanged)

	for nodes.Next() {
		node := nodes.Node().(graph.NodeDevice)
		if node.Device().InUse() && !network.IsLocalDevice(node) {
			multisrv.createApiAndOtaServers(MeshNodeId(node.ID()), network)
		}
	}
	return &multisrv
}
