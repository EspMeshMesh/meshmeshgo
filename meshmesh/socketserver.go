package meshmesh

import (
	"errors"
	"fmt"
	"net"

	"golang.org/x/exp/slices"
	"leguru.net/m/v2/graph"
	"leguru.net/m/v2/logger"
	"leguru.net/m/v2/utils"
)

var _allStats *EspApiStats

const (
	fixedApiRemotePort int = 6053
	fixedOtaRemotePort int = 3232
)

type SocketServer struct {
	Address       MeshNodeId
	Clients       []*ConnectionPathBridge
	listener      net.Listener
	listenAddress string
	listnerClosed func(*SocketServer)
	network       *graph.Network
	hasShutdown   bool
}

func (s *SocketServer) connectionFactory(remotePort int) ConnectionPathBridgeDriver {
	var client ConnectionPathBridgeDriver = nil

	if remotePort == fixedApiRemotePort {
		client = NewApiConnection()
	} else {
		client = NewOtaConnection()
	}

	return client
}

func (s *SocketServer) GetListenAddress() string {
	return s.listenAddress
}

func (s *SocketServer) ClientClosedCb(client *ConnectionPathBridge) {
	// Remove client from clients list
	client.Dispose()
	idx := slices.Index(s.Clients, client)
	if idx >= 0 {
		s.Clients = append(s.Clients[:idx], s.Clients[idx+1:]...)
	}
	logger.WithFields(logger.Fields{"nodeId": utils.FmtNodeId(int64(s.Address)), "address": s.listenAddress}).Debug("SocketServer.ClientClosedCb. Removed client from clients list")
}

func (s *SocketServer) ListenAndServe(serialProxy *ConnectedPath2Serial, remotePort int) {
	for {
		socket, err := s.listener.Accept()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				logger.Error(err)
			}
			break
		}

		logger.WithFields(logger.Fields{"nodeId": utils.FmtNodeId(int64(s.Address)), "address": s.listenAddress, "active": len(s.Clients)}).Debug("ServerSocket.ListenAndServe: connection accepted")

		driver := s.connectionFactory(remotePort)
		client, err := NewConnectionPathBridge(socket, serialProxy, s.network, s.Address, remotePort, driver, s.ClientClosedCb)
		if err == nil {
			s.Clients = append(s.Clients, client)
		} else {
			logger.Error(err)
			socket.Close()
		}
	}
	s.ShutDown()
}

func (s *SocketServer) ShutDown() {
	if s.hasShutdown {
		return
	}
	logger.WithFields(logger.Fields{"nodeId": utils.FmtNodeId(int64(s.Address)), "address": s.listenAddress}).Info("SocketServer.ShutDown. Shutting down listener and all clients")
	s.hasShutdown = true
	s.listener.Close()
	for _, client := range s.Clients {
		client.close()
	}
	if s.listnerClosed != nil {
		s.listnerClosed(s)
	}
	s.hasShutdown = true
}

func NewSocketServer(serialProxy *ConnectedPath2Serial, network *graph.Network, address MeshNodeId, config *ServerApiConfig, listenerClosed func(*SocketServer)) (*SocketServer, error) {
	var bindAddress string = config.BindAddress
	if config.BindAddress == "" || config.BindAddress == "dynamic" {
		bindAddress = utils.FmtNodeIdHass(int64(address))
	}

	bindPort := config.BindPort
	if config.BindPort <= 0 {
		bindPort = utils.HashString(utils.FmtNodeId(int64(address)), config.SizeOfPortsPool) + config.BasePortOffset
	}

	server := SocketServer{Address: address, network: network, listnerClosed: listenerClosed}
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
