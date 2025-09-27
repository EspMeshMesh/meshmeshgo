package meshmesh

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"golang.org/x/exp/slices"
	"leguru.net/m/v2/graph"
	"leguru.net/m/v2/logger"
	"leguru.net/m/v2/utils"
)

var _allStats *EspApiStats

const (
	esphomeapiWaitPacketHead int = 0
	esphomeapiWaitPacketSize int = 1
	esphomeapiWaitPacketData int = 3
)

const (
	fixedApiRemotePort int = 6053
	fixedOtaRemotePort int = 3232
)

type NetworkConnection interface {
	MeshProtocol() *ConnPathConnection
	FinishHandshake(result bool)
	ForwardData(data []byte) error
	Close()
}

type NetworkConnectionStruct struct {
	Stats           *EspApiConnectionStats
	tmpBuffer       *bytes.Buffer
	inBuffer        *bytes.Buffer
	socketOpen      bool
	socket          net.Conn
	socketWaitGroup sync.WaitGroup
	meshprotocol    *ConnPathConnection
	reqAddress      MeshNodeId
	reqPort         int
	debugThisNode   bool
	timeout         time.Time
	clientClosed    func(client NetworkConnection)
}

func (c *NetworkConnectionStruct) MeshProtocol() *ConnPathConnection {
	return c.meshprotocol
}

func NewNetworkConnectionStruct(socket net.Conn, serial *SerialConnection, addr MeshNodeId, port int, closedCb func(NetworkConnection)) NetworkConnectionStruct {
	return NetworkConnectionStruct{
		meshprotocol: NewConnPathConnection(serial),
		socketOpen:   true,
		socket:       socket,
		tmpBuffer:    bytes.NewBuffer([]byte{}),
		inBuffer:     bytes.NewBuffer([]byte{}),
		timeout:      time.Now(),
		clientClosed: closedCb,
		Stats:        _allStats.Stats(addr),
	}
}

type ServerApi struct {
	Address       MeshNodeId
	Clients       []NetworkConnection
	listener      net.Listener
	listenAddress string
}

func (s *ServerApi) GetListenAddress() string {
	return s.listenAddress
}

func (s *ServerApi) HandleConnectedPathReply(v *ConnectedPathApiReply) bool {
	handled := false
	for _, client := range s.Clients {
		if client.MeshProtocol().handle == v.Handle {
			handled = true
			if v.Command == connectedPathSendDataRequest {
				if len(v.Data) > 0 {
					err := client.ForwardData(v.Data)
					if err != nil {
						logger.Printf("HandleConnectedPathReply: ForwardData error on handle %d.", v.Handle)
						client.Close()
					}
				}
			} else {
				oldConnState := client.MeshProtocol().connState
				client.MeshProtocol().HandleIncomingReply(v)
				if oldConnState != client.MeshProtocol().connState {
					if client.MeshProtocol().connState == connPathConnectionStateActive {
						client.FinishHandshake(true)
					}
					if client.MeshProtocol().connState == connPathConnectionStateInvalid {
						client.Close()
					}
				}
			}
		}
	}
	return handled
}

func (s *ServerApi) ClientClosedCb(client NetworkConnection) {
	// Remove client from clients list
	idx := slices.Index(s.Clients, client)
	if idx >= 0 {
		s.Clients = append(s.Clients[:idx], s.Clients[idx+1:]...)
	}
	logger.WithFields(logger.Fields{"handle": client.MeshProtocol().handle}).Info("Closed EspHomeApi connection")
}

func (s *ServerApi) ListenAndServe(serial *SerialConnection, remotePort int) {
	for {
		socket, err := s.listener.Accept()
		if err != nil {
			logger.Error(err)
			continue
		}

		logger.WithFields(logger.Fields{"nodeId": s.Address, "active": len(s.Clients)}).Debug("EspHome connection accepted")

		if remotePort == fixedApiRemotePort {
			client, err := NewApiConnection(socket, serial, s.Address, remotePort, s.ClientClosedCb)
			if err != nil {
				logger.Error(err)
				socket.Close()
			} else {
				s.Clients = append(s.Clients, client)
				logger.WithFields(logger.Fields{"nodeId": utils.FmtNodeId(int64(s.Address)), "clients": len(s.Clients)}).Debug("Added new client")
			}
		} else {
			client, err := NewOtaConnection(socket, serial, s.Address, remotePort, s.ClientClosedCb)
			if err != nil {
				logger.Error(err)
				socket.Close()
			} else {
				s.Clients = append(s.Clients, client)
				logger.WithFields(logger.Fields{"nodeId": utils.FmtNodeId(int64(s.Address)), "clients": len(s.Clients)}).Debug("Added new client")
			}
		}

	}
}

func (s *ServerApi) CloseConnections() {
	for _, client := range s.Clients {
		client.Close()
	}
}

func (s *ServerApi) ShutDown() {
	s.listener.Close()
}

func NewServerApi(serial *SerialConnection, address MeshNodeId, config *ServerApiConfig) (*ServerApi, error) {
	var bindAddress string = config.BindAddress
	if config.BindAddress == "" || config.BindAddress == "dynamic" {
		bindAddress = utils.FmtNodeIdHass(int64(address))
	}

	bindPort := config.BindPort
	if config.BindPort <= 0 {
		bindPort = utils.HashString(utils.FmtNodeId(int64(address)), config.SizeOfPortsPool) + config.BasePortOffset
	}

	server := ServerApi{Address: address}
	server.listenAddress = fmt.Sprintf("%s:%d", bindAddress, bindPort)
	listener, err := net.Listen("tcp4", server.listenAddress)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	logger.WithFields(logger.Fields{"node": utils.FmtNodeId(int64(address)), "bind": server.listenAddress}).Debug("Start listening on port for node connection")
	server.listener = listener
	go server.ListenAndServe(serial, config.RemotePort)
	return &server, nil
}

func (m *MultiServerApi) handleUnhandledReply(v *ConnectedPathApiReply) {
	logger.WithFields(logger.Fields{"cmd": v.Command, "handle": v.Handle}).
		Error("handleUnhandledReply: Connection not found for this handle")
}

func (m *MultiServerApi) HandleConnectedPathReply(v *ConnectedPathApiReply) {
	var handled bool = false
	for _, server := range m.Servers {
		handled = server.HandleConnectedPathReply(v)
		if handled {
			break
		}
	}
	if !handled {
		m.handleUnhandledReply(v)
	}
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

func (m *MultiServerApi) MainNetworkChanged() {
	nodes := graph.GetMainNetwork().Nodes()
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
				server, err := NewServerApi(m.serial, MeshNodeId(node.ID()), &m.config)
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

type ServerApiConfig struct {
	BindAddress     string
	BindPort        int
	RemotePort      int
	BasePortOffset  int
	SizeOfPortsPool int
}

type MultiServerApi struct {
	espApiStats *EspApiStats
	serial      *SerialConnection
	config      ServerApiConfig
	Servers     []*ServerApi
}

func NewMultiServerApi(serial *SerialConnection, config ServerApiConfig) *MultiServerApi {
	_allStats = NewEspApiStats()
	multisrv := MultiServerApi{serial: serial, espApiStats: _allStats, config: config}
	SendClearConnections(serial)
	multisrv.serial.ConnPathFn = multisrv.HandleConnectedPathReply
	localNode := serial.LocalNode

	nodes := graph.GetMainNetwork().Nodes()
	graph.AddMainNetworkChangedCallback(multisrv.MainNetworkChanged)

	for nodes.Next() {
		node := nodes.Node().(graph.NodeDevice)
		if node.Device().InUse() && node.ID() != int64(localNode) {
			configApi := config
			configApi.RemotePort = fixedApiRemotePort

			server, err := NewServerApi(serial, MeshNodeId(node.ID()), &configApi)
			if err != nil {
				log.Error(err)
			} else {
				multisrv.Servers = append(multisrv.Servers, server)
			}

			configOta := config
			configOta.BindPort = fixedOtaRemotePort
			configOta.RemotePort = fixedOtaRemotePort
			serverOta, err := NewServerApi(serial, MeshNodeId(node.ID()), &configOta)
			if err != nil {
				log.Error(err)
			} else {
				multisrv.Servers = append(multisrv.Servers, serverOta)
			}
		}
	}
	return &multisrv
}
