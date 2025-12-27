package meshmesh

import (
	"github.com/charmbracelet/log"
	"golang.org/x/exp/slices"
	"leguru.net/m/v2/graph"
	"leguru.net/m/v2/logger"
	"leguru.net/m/v2/utils"
)

func (m *MultiSocketServer) Stats() *EspApiStats {
	return m.espApiStats
}

func (m *MultiSocketServer) PrintStats() {
	m.espApiStats.PrintStats()
}

func (m *MultiSocketServer) ShutdownServer(addr MeshNodeId) {
	for _, server := range m.Servers {
		if server.Address == addr {
			server.ShutDown()
		}
	}
}

func (m *MultiSocketServer) StarPathProtocol(starpath *StarPath) {
	m.starPathNetwork = starpath.network
	m.starPathNetwork.AddNetworkChangedCallback(m.networkChanged)
	m.networkChanged(m.starPathNetwork)
}

func (m *MultiSocketServer) serverAddressExists(nodeId MeshNodeId) bool {
	for _, server := range m.Servers {
		if server.Address == nodeId {
			return true
		}
	}
	return false
}

func (m *MultiSocketServer) listenerClosed(server *SocketServer) {
	idx := slices.Index(m.Servers, server)
	if idx >= 0 {
		m.Servers = append(m.Servers[:idx], m.Servers[idx+1:]...)
	}
	logger.WithFields(logger.Fields{"server": utils.FmtNodeId(int64(server.Address))}).Debug("MultiSocketServer.listenerClosed: removing server")
}

func (m *MultiSocketServer) createApiAndOtaServers(nodeId MeshNodeId, network *graph.Network) {
	configApi := m.config
	configApi.RemotePort = fixedApiRemotePort

	server, err := NewSocketServer(m.serialProxy, network, nodeId, &configApi, m.listenerClosed)
	if err != nil {
		log.Error(err)
	} else {
		m.Servers = append(m.Servers, server)
	}

	configOta := m.config
	configOta.BindPort = fixedOtaRemotePort
	configOta.RemotePort = fixedOtaRemotePort
	serverOta, err := NewSocketServer(m.serialProxy, network, nodeId, &configOta, m.listenerClosed)
	if err != nil {
		log.Error(err)
	} else {
		m.Servers = append(m.Servers, serverOta)
	}
}

func (m *MultiSocketServer) networkChanged(network *graph.Network) {
	nodes := network.Nodes()
	for nodes.Next() {
		node := nodes.Node().(graph.NodeDevice)
		wantServer := node.Device().InUse() && !network.IsLocalDevice(node) && !node.Device().DeepSleep()
		hasServer := m.serverAddressExists(MeshNodeId(node.ID()))

		if wantServer && !hasServer {
			logger.WithFields(logger.Fields{"nodeId": utils.FmtNodeId(int64(node.ID())), "network": network.NetworkId()}).Debug("MultiSocketServer.networkChanged: adding new node server")
			m.createApiAndOtaServers(MeshNodeId(node.ID()), network)
		}
	}

	oldnodes := make([]MeshNodeId, 0)
	for _, server := range m.Servers {
		addr := server.Address
		if m.mainNetwork.Node(int64(addr)) == nil && m.starPathNetwork.Node(int64(addr)) == nil {
			oldnodes = append(oldnodes, addr)
		}
	}
	for _, addr := range oldnodes {
		m.ShutdownServer(addr)
	}
}

type ServerApiConfig struct {
	BindAddress     string
	BindPort        int
	RemotePort      int
	BasePortOffset  int
	SizeOfPortsPool int
}

type MultiSocketServer struct {
	espApiStats     *EspApiStats
	serialProxy     *ConnectedPath2Serial
	starPathNetwork *graph.Network
	mainNetwork     *graph.Network
	config          ServerApiConfig
	Servers         []*SocketServer
}

func NewMultiSocketServer(serialProxy *ConnectedPath2Serial, config ServerApiConfig) *MultiSocketServer {
	_allStats = NewEspApiStats()
	multisrv := MultiSocketServer{serialProxy: serialProxy, espApiStats: _allStats, config: config}
	serialProxy.ClearConnections()

	multisrv.mainNetwork = graph.GetMainNetwork()
	multisrv.mainNetwork.AddNetworkChangedCallback(multisrv.networkChanged)
	multisrv.networkChanged(multisrv.mainNetwork)
	return &multisrv
}
