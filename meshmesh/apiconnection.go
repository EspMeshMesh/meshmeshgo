package meshmesh

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"time"

	"leguru.net/m/v2/graph"
	"leguru.net/m/v2/logger"
	"leguru.net/m/v2/utils"
)

type ApiConnection struct {
	NetworkConnectionStruct
	inState     int
	inAwaitSize int
}

func (client *ApiConnection) forward(lastbyte byte) {
	client.inBuffer.WriteByte(lastbyte)
	switch client.inState {
	case esphomeapiWaitPacketHead:
		if lastbyte == 0x00 {
			client.inState = esphomeapiWaitPacketSize
		} else {
			client.inBuffer.Reset()
		}
	case esphomeapiWaitPacketSize:
		client.inAwaitSize = int(lastbyte) + 3
		client.inState = esphomeapiWaitPacketData
	default:
		if client.inBuffer.Len() == client.inAwaitSize {
			client.inState = esphomeapiWaitPacketHead
			logger.WithField("handle", client.meshprotocol.handle).
				Trace(fmt.Sprintf("HA-->SE: %s", hex.EncodeToString(client.inBuffer.Bytes())))
			err := client.meshprotocol.SendData(client.inBuffer.Bytes())
			client.Stats.SentBytes(client.inBuffer.Len())
			if err != nil {
				logger.Log().Error(fmt.Sprintf("Error writng on socket: %s", err.Error()))
			}
			client.inBuffer.Reset()
		}
	}
}

func (client *ApiConnection) flushBuffer() {
	if client.tmpBuffer.Len() > 0 {
		_b := client.tmpBuffer.Bytes()
		for i := 0; i < len(_b); i++ {
			client.forward(_b[i])
		}
	}
}

func (client *ApiConnection) SetClosedCallback(cb func(client NetworkConnection)) {
	client.clientClosed = cb
}

func (client *ApiConnection) Close() {
	client.socketOpen = false
	client.socket.Close()
	client.socketWaitGroup.Wait()
	client.Stats.Stop()
	client.meshprotocol.Disconnect()
	client.clientClosed(client)
	logger.Log().Debug("ApiConnection.Close")
}

func (client *ApiConnection) CheckTimeout() {
	for {
		if !client.socketOpen {
			break
		}
		if client.meshprotocol.connState == connPathConnectionStateInit || client.meshprotocol.connState == connPathConnectionStateHandshakeStarted {
			if time.Since(client.timeout).Milliseconds() > 3000 {
				logger.Error(fmt.Sprintf("Closing connection beacuse timeout after %dms in connPathConnectionStateInit for handle %d", time.Since(client.timeout).Milliseconds(), client.meshprotocol.handle))
				client.Close()
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	logger.Debug("ApiConnection.CheckTimeout exited")
}

func (client *ApiConnection) Read() {
	var err error

	for {
		var buffer = make([]byte, 1)
		_, err = client.socket.Read(buffer)
		client.Stats.ReceivedBytes(1)

		if err == nil {
			switch client.meshprotocol.connState {
			case connPathConnectionStateHandshakeStarted:
				// FIXME check for if buffer grown outside limits
				client.tmpBuffer.WriteByte(buffer[0])
			case connPathConnectionStateActive:
				// FIXME handle error
				client.forward(buffer[0])
			default:
				logger.WithField("state", client.meshprotocol.connState).
					Error(fmt.Errorf("readed data while in wrong connection state %d", client.meshprotocol.connState))
			}
		} else {
			if errors.Is(err, net.ErrClosed) {
				// Log other errors than net.ErrClosed
			} else {
				logger.WithFields(logger.Fields{"handle": client.meshprotocol.handle, "err": err}).Error("ApiConnection.Read exit with error")
			}
			break
		}
	}

	client.socketWaitGroup.Done()
	if client.socketOpen {
		logger.Log().Debug("ApiConnection.Read close socket endpoint")
		client.Close()
	}
}

func (client *ApiConnection) ForwardData(data []byte) error {
	logger.WithFields(logger.Fields{
		"handle": client.meshprotocol.handle,
		"meshid": utils.FmtNodeId(int64(client.reqAddress)),
		"len":    len(data),
		"data":   utils.EncodeToHexEllipsis(data, 10),
	}).Trace("SE-->HA")
	n, err := client.socket.Write(data)
	if err != nil {
		return err
	}

	if n < len(data) {
		return errors.New("socket can't receive all bytes")
	}

	return nil
}

func NewApiConnection(socket net.Conn, serial *SerialConnection, network *graph.Network, addr MeshNodeId, port int, closedCb func(NetworkConnection)) (*ApiConnection, error) {
	if !serial.IsConnected() {
		return nil, errors.New("serial is not open")
	}

	client := &ApiConnection{
		NetworkConnectionStruct: NewNetworkConnectionStruct(socket, serial, network, addr, port, closedCb),
	}

	err := client.startHandshake(addr, port)
	if err != nil {
		return nil, err
	}

	client.socketWaitGroup.Add(1)
	go client.Read()
	go client.CheckTimeout()

	return client, nil
}
