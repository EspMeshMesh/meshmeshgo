package meshmesh

import (
	"errors"
	"strconv"
	"strings"

	"leguru.net/m/v2/graph"
	"leguru.net/m/v2/logger"
	"leguru.net/m/v2/utils"
)

const connectedPathOpenConnectionRequest uint8 = 1
const connectedPathSendDataNackReply uint8 = 4
const connectedPathSendDataRequest uint8 = 5
const connectedPathOpenConnectionAck uint8 = 6
const connectedPathOpenConnectionNack uint8 = 7
const connectedPathDisconnectRequest uint8 = 8

// const connectedPathSendDataError uint8 = 9
const connectedPathClearConnections uint8 = 10

const (
	connPathConnectionStateInit uint8 = iota
	connPathConnectionStateHandshakeStarted
	connPathConnectionStateHandshakeFailed
	connPathConnectionStateActive
	connPathConnectionStateInvalid
)

type ConnPathConnection struct {
	serialProxy                 *ConnectedPath2Serial
	serialDataAvailableCallback func(data []byte)
	connectionActiveCallback    func()
	connectionInvalidCallback   func()
	connState                   uint8
	handle                      uint16
	sequence                    uint16
	network                     *graph.Network
}

func ParseAddress(address string) (MeshNodeId, error) {
	fields := strings.Split(address, ".")
	if len(fields) != 4 {
		return 0, errors.New("invalid address string")
	}

	var err error
	addr := make([]byte, 4)
	for i, field := range fields {
		var data int
		data, err = strconv.Atoi(field)
		if err != nil {
			break
		}
		addr[i] = byte(data)
	}

	if err != nil {
		return 0, err
	} else {
		return (MeshNodeId(addr[1]) << 16) + (MeshNodeId(addr[2]) << 8) + MeshNodeId(addr[3]), nil
	}
}

func (client *ConnPathConnection) getNextSequence() uint16 {
	client.sequence += 1
	if client.sequence == 0 {
		client.sequence = 1
	}
	return client.sequence
}

func (client *ConnPathConnection) SendData(data []byte) error {
	err := client.serialProxy.sendFrame(connectedPathSendDataRequest, client.handle, client.getNextSequence(), data)
	return err
}

func (client *ConnPathConnection) SendDataNack() error {
	return client.serialProxy.sendFrame(connectedPathSendDataNackReply, client.handle, client.getNextSequence(), []byte{})
}

func (client *ConnPathConnection) OpenConnectionAsync2(textaddr string, port uint16) error {
	addr, err := ParseAddress(textaddr)
	if err != nil {
		return err
	}

	return client.OpenConnectionAsync(addr, port)
}

func (client *ConnPathConnection) OpenConnectionAsync(addr MeshNodeId, port uint16) error {
	logger.WithFields(logger.Fields{"addr": utils.FmtNodeId(int64(addr)), "port": port, "handle": client.handle}).
		Debug("ConnPathConnection.OpenConnectionAsync")

	network := client.network
	device, err := network.GetNodeDevice(int64(addr))
	if err != nil {
		return err
	}
	_path, _, err := network.GetPath(device)
	if err != nil {
		return err
	}
	if len(_path) == 1 {
		return errors.New("speak with local node is not yet supported")
	}

	_path = _path[1:]
	path := make([]int32, len(_path))
	for i, item := range _path {
		path[i] = int32(item)
	}

	client.connState = connPathConnectionStateHandshakeStarted
	err = client.serialProxy.sendOpenConnectionRequest(client.handle, client.getNextSequence(), port, path)
	return err
}

func (client *ConnPathConnection) Disconnect() {

	// Only send disconnect request if the coordinator connection is active
	if client.connState != connPathConnectionStateActive {
		return
	}

	logger.WithFields(logger.Fields{"handle": client.handle, "connState": client.connState}).Debug("Sending Disconnect request")
	client.serialProxy.sendFrame(connectedPathDisconnectRequest, client.handle, client.getNextSequence(), []byte{})
	client.invalidateConnection()
}

func (client *ConnPathConnection) handleIncomingSendDataRequest(v *ConnectedPathApiReply) {
	if len(v.Data) > 0 {
		client.serialDataAvailableCallback(v.Data)
	}
}

func (client *ConnPathConnection) handleIncomingOpenConnAck(_ *ConnectedPathApiReply) {
	if client.connState != connPathConnectionStateHandshakeStarted {
		logger.Error("handleIncomingOpenConnAck received while not in handshake state")
		client.invalidateConnection()
	} else {
		logger.WithField("handle", client.handle).Debug("Accpeted connection")
		client.connState = connPathConnectionStateActive
		if client.connectionActiveCallback != nil {
			client.connectionActiveCallback()
		}
	}
}

func (client *ConnPathConnection) handleIncomingOpenConnNack(v *ConnectedPathApiReply) {
	logger.WithFields(logger.Fields{"handle": v.Handle}).Error("nack during opening connection")
	client.invalidateConnection()
}

func (client *ConnPathConnection) handleIncomingSerialPacket(v *ConnectedPathApiReply) {
	switch v.Command {
	case connectedPathSendDataRequest:
		client.handleIncomingSendDataRequest(v)
	case connectedPathOpenConnectionAck:
		client.handleIncomingOpenConnAck(v)
	case connectedPathOpenConnectionNack:
		client.handleIncomingOpenConnNack(v)
	case connectedPathSendDataNackReply:
		logger.WithField("handle", v.Handle).Error("HandleIncomingReply: SendDataNack")
		client.invalidateConnection()
	case connectedPathDisconnectRequest:
		logger.WithField("handle", v.Handle).Debug("HandleIncomingReply: DisconnectRequest")
		client.invalidateConnection()
	default:
		logger.WithFields(logger.Fields{"handle": v.Handle, "reply": v.Command}).
			Error("HandleIncomingReply: unknow command reply received", v.Command, v.Handle)
	}
}

func (client *ConnPathConnection) invalidateConnection() {
	if client.connState != connPathConnectionStateInvalid {
		client.connState = connPathConnectionStateInvalid
		if client.connectionInvalidCallback != nil {
			client.connectionInvalidCallback()
		}
	}
}

func (client *ConnPathConnection) SetSerialDataAvailableCallback(callback func(data []byte)) {
	client.serialDataAvailableCallback = callback
}

func (client *ConnPathConnection) SetConnectionActiveCallback(callback func()) {
	client.connectionActiveCallback = callback
}

func (client *ConnPathConnection) SetConnectionInvalidCallback(callback func()) {
	client.connectionInvalidCallback = callback
}

func (client *ConnPathConnection) Dispose() {
	client.serialProxy.RemovePacketReceivedCallback(client.handle)
}

func NewConnPathConnection(serialProxy *ConnectedPath2Serial, network *graph.Network) *ConnPathConnection {

	conn := &ConnPathConnection{
		serialProxy: serialProxy,
		network:     network,
		handle:      serialProxy.GetNextSerialHandle(),
		connState:   connPathConnectionStateInit,
	}

	serialProxy.AddPacketReceivedCallback(conn.handle, conn.handleIncomingSerialPacket)
	return conn
}
