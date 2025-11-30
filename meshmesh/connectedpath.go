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
	//address   MeshNodeId
	connState uint8
	serial    *SerialConnection
	handle    uint16
	sequence  uint16
	network   *graph.Network
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
	err := client.serial.SendApi(ConnectedPathApiRequest{
		Protocol: meshmeshProtocolConnectedPath,
		Command:  connectedPathSendDataRequest,
		Handle:   client.handle,
		Dummy:    0,
		Sequence: client.getNextSequence(),
		DataSize: uint16(len(data)),
		Data:     data,
	})

	return err
}

func SendSendDataNack(serial *SerialConnection, handle uint16) error {
	err := serial.SendApi(ConnectedPathApiRequest{
		Protocol: meshmeshProtocolConnectedPath,
		Command:  connectedPathSendDataNackReply,
		Handle:   handle,
		Dummy:    0,
		Sequence: 0,
		DataSize: 0,
		Data:     []byte{},
	})

	return err
}

func SendClearConnections(serial *SerialConnection) error {
	err := serial.SendApi(ConnectedPathApiRequest{
		Protocol: meshmeshProtocolConnectedPath,
		Command:  connectedPathClearConnections,
		Handle:   0,
		Dummy:    0,
		Sequence: 0,
		DataSize: 0,
		Data:     []byte{},
	})

	return err
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
	err = client.serial.SendApi(
		ConnectedPathApiRequest2{
			Protocol: meshmeshProtocolConnectedPath,
			Command:  connectedPathOpenConnectionRequest,
			Handle:   client.handle,
			Dummy:    0,
			Sequence: client.getNextSequence(),
			DataSize: uint16(len(path)*4 + 3),
			Port:     port,
			PathLen:  uint8(len(path)),
			Path:     path,
		},
	)

	return err
}

func (client *ConnPathConnection) Disconnect() {

	// Only send disconnect request if the coordinator connection is active
	if client.connState != connPathConnectionStateActive {
		return
	}

	logger.WithFields(logger.Fields{"handle": client.handle, "connState": client.connState}).Debug("Sending Disconnect request")
	client.serial.SendApi(ConnectedPathApiRequest{
		Protocol: meshmeshProtocolConnectedPath,
		Command:  connectedPathDisconnectRequest,
		Handle:   client.handle,
		Dummy:    0,
		Sequence: client.getNextSequence(),
		DataSize: 0,
		Data:     []byte{},
	})
	client.connState = connPathConnectionStateInvalid
}

func (client *ConnPathConnection) handleIncomingOpenConnAck(_ *ConnectedPathApiReply) {
	if client.connState != connPathConnectionStateHandshakeStarted {
		client.connState = connPathConnectionStateInvalid
		logger.Error("handleIncomingOpenConnAck received while not in handshake state")
	} else {
		logger.WithField("handle", client.handle).Debug("Accpeted connection")
		client.connState = connPathConnectionStateActive

	}
}

func (client *ConnPathConnection) handleIncomingOpenConnNack(v *ConnectedPathApiReply) {
	logger.WithFields(logger.Fields{"handle": v.Handle}).Error("nack during opening connection")
	client.connState = connPathConnectionStateInvalid
}

func (client *ConnPathConnection) HandleIncomingReply(v *ConnectedPathApiReply) {
	logger.WithFields(logger.Fields{"handle": v.Handle, "reply": v.Command}).Debug("HandleIncomingReply")
	switch v.Command {
	case connectedPathOpenConnectionAck:
		client.handleIncomingOpenConnAck(v)
	case connectedPathOpenConnectionNack:
		client.handleIncomingOpenConnNack(v)
	case connectedPathSendDataNackReply:
		logger.WithField("handle", v.Handle).Error("HandleIncomingReply: SendDataNack")
		client.connState = connPathConnectionStateInvalid
	case connectedPathDisconnectRequest:
		logger.WithField("handle", v.Handle).Debug("HandleIncomingReply: DisconnectRequest")
		client.connState = connPathConnectionStateInvalid
	default:
		logger.WithFields(logger.Fields{"handle": v.Handle, "reply": v.Command}).
			Error("HandleIncomingReply: unknow command reply received", v.Command, v.Handle)
	}
}

func NewConnPathConnection(serial *SerialConnection, network *graph.Network) *ConnPathConnection {

	conn := &ConnPathConnection{
		serial:    serial,
		network:   network,
		handle:    serial.GetNextHandle(),
		connState: connPathConnectionStateInit,
	}
	return conn

}
