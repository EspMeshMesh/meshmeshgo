package meshmesh

import "leguru.net/m/v2/logger"

type ConnectedPathPacketReceivedCallback struct {
	Handle   uint16
	Callback func(packet *ConnectedPathApiReply)
}

type ConnectedPath2Serial struct {
	serial                 *SerialConnection
	packetReceivedCallback []ConnectedPathPacketReceivedCallback
}

func (conn *ConnectedPath2Serial) handleIncomingFrame(data any) {
	cp, ok := data.(ConnectedPathApiReply)
	if !ok {
		logger.Log().Error("Can't decode incoming connectedpath packet 1/2")
		return
	}

	for _, callback := range conn.packetReceivedCallback {
		if callback.Handle == cp.Handle {
			callback.Callback(&cp)
			return
		}
	}

	logger.WithFields(logger.Fields{"handle": cp.Handle}).Error("No callback found for connectedpath packet")
}

func (conn *ConnectedPath2Serial) sendFrame(command uint8, handle uint16, sequence uint16, data []byte) error {
	err := conn.serial.SendApi(ConnectedPathApiRequest{
		Protocol: meshmeshProtocolConnectedPath,
		Command:  command,
		Handle:   handle,
		Dummy:    0,
		Sequence: sequence,
		DataSize: uint16(len(data)),
		Data:     data,
	})
	return err
}

func (conn *ConnectedPath2Serial) sendOpenConnectionRequest(handle uint16, sequence uint16, port uint16, path []int32) error {
	err := conn.serial.SendApi(ConnectedPathApiRequest2{
		Protocol: meshmeshProtocolConnectedPath,
		Command:  connectedPathOpenConnectionRequest,
		Handle:   handle,
		Dummy:    0,
		Sequence: sequence,
		DataSize: uint16(len(path)*4 + 3),
		Port:     port,
		PathLen:  uint8(len(path)),
		Path:     path,
	})

	return err
}

func (conn *ConnectedPath2Serial) IsSerialConnected() bool {
	return conn.serial.IsConnected()
}

func (client *ConnectedPath2Serial) ClearConnections() error {
	return client.sendFrame(connectedPathClearConnections, 0, 0, []byte{})
}

func (conn *ConnectedPath2Serial) IsEsp8266() bool {
	return conn.serial.isEsp8266
}

func (conn *ConnectedPath2Serial) TxOneByteMs() int {
	return conn.serial.txOneByteMs
}

func (conn *ConnectedPath2Serial) AddPacketReceivedCallback(handle uint16, callback func(packet *ConnectedPathApiReply)) {
	conn.packetReceivedCallback = append(conn.packetReceivedCallback, ConnectedPathPacketReceivedCallback{Handle: handle, Callback: callback})
}

func (conn *ConnectedPath2Serial) RemovePacketReceivedCallback(handle uint16) {
	for i, callback := range conn.packetReceivedCallback {
		if callback.Handle == handle {
			conn.packetReceivedCallback = append(conn.packetReceivedCallback[:i], conn.packetReceivedCallback[i+1:]...)
		}
	}
}

func (conn *ConnectedPath2Serial) GetNextSerialHandle() uint16 {
	return conn.serial.GetNextHandle()
}

func NewConnectedPath2Serial(serial *SerialConnection) *ConnectedPath2Serial {
	conn := &ConnectedPath2Serial{
		serial: serial,
	}
	serial.AddFrameReceivedCallback(connectedPathApiReply, 0, conn.handleIncomingFrame)
	return conn
}
