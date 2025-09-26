package meshmesh

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"leguru.net/m/v2/logger"
	"leguru.net/m/v2/utils"
)

type OtaConnection struct {
	NetworkConnectionStruct
}

func (client *OtaConnection) startHandshake(addr MeshNodeId, port int) error {
	client.reqAddress = addr
	client.reqPort = port
	err := client.meshprotocol.OpenConnectionAsync(addr, uint16(port))
	if err == nil {
		client.Stats.Start()
		if addr == MeshNodeId(0) {
			client.debugThisNode = true
			logger.WithFields(logger.Fields{"id": fmt.Sprintf("%02X", addr)}).Info("startHandshake and debug for node")
		}
	}
	return err
}

func (client *OtaConnection) FinishHandshake(result bool) {
	logger.WithField("res", result).Debug("OtaConnection.FinishHandshake")
	if !result {
		logger.WithFields(logger.Fields{"addr": client.reqAddress, "port": client.reqPort, "err": nil}).
			Warning("OtaConnection.FinishHandshake failed")
	} else {
		logger.WithFields(logger.Fields{"addr": client.reqAddress, "port": client.reqPort, "handle": client.meshprotocol.handle}).
			Info("OtaConnection.FinishHandshake OpenConnection succesfull")
		client.flushBuffer(client.tmpBuffer)
		client.Stats.GotHandle(client.meshprotocol.handle)
	}
}

func (client *OtaConnection) flushBuffer(buffer *bytes.Buffer) {
	if buffer.Len() > 0 {
		logger.WithFields(logger.Fields{"handle": client.meshprotocol.handle, "len": buffer.Len()}).
			Trace(fmt.Sprintf("flushBuffer: HA-->SE: %s", utils.EncodeToHexEllipsis(buffer.Bytes(), 32)))

		chunks := (buffer.Len()-1)/512 + 1

		for i := 0; i < chunks; i++ {
			chunk := buffer.Next(512)
			err := client.meshprotocol.SendData(chunk)
			if err != nil {
				logger.Log().Error(fmt.Sprintf("Error writing on socket: %s", err.Error()))
			}
			if client.meshprotocol.serial.isEsp8266 {
				sleepTime := client.meshprotocol.serial.txOneByteMs * (len(chunk) * 25)
				time.Sleep(time.Duration(sleepTime) * time.Microsecond)
			}
		}

		//client.meshprotocol.SendData([]byte{})

		client.Stats.SentBytes(buffer.Len())
		buffer.Reset()
	}
}

func (client *OtaConnection) SetClosedCallback(cb func(client NetworkConnection)) {
	client.clientClosed = cb
}

func (client *OtaConnection) Close() {
	client.socketOpen = false
	client.socket.Close()
	logger.Log().Debug("OtaConnection.Close close socket endpoint")
	client.socketWaitGroup.Wait()
	client.Stats.Stop()
	client.meshprotocol.Disconnect()
	client.clientClosed(client)
}

func (client *OtaConnection) CheckTimeout() {
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

func (client *OtaConnection) Read() {
	var n int
	var err error

	for {
		var buffer = make([]byte, 1)
		client.socket.SetReadDeadline(time.Now().Add(10 * time.Millisecond))

		n, err = client.socket.Read(buffer)
		client.Stats.ReceivedBytes(n)

		/*if n > 0 {
			logger.WithFields(logger.Fields{"handle": client.socket.RemoteAddr().String(), "n": n, "bytes": utils.EncodeToHexEllipsis(buffer, 10)}).Debug("OtaConnection.Read")
		}*/

		if err == io.EOF {
			logger.WithFields(logger.Fields{"handle": client.socket.RemoteAddr().String()}).Warn("OtaConnection.Read connection closed by peer")
			break
		}

		if err != nil {
			if !errors.Is(err, os.ErrDeadlineExceeded) {
				logger.WithFields(logger.Fields{"handle": client.socket.RemoteAddr().String(), "err": err}).Error("OtaConnection.Read error")
				break
			}
		}

		if n > 0 {
			switch client.meshprotocol.connState {
			case connPathConnectionStateHandshakeStarted:
				// FIXME check for if buffer grown outside limits
				client.tmpBuffer.WriteByte(buffer[0])
			case connPathConnectionStateActive:
				// FIXME handle error
				client.inBuffer.WriteByte(buffer[0])
			default:
				logger.WithField("state", client.meshprotocol.connState).
					Error(fmt.Errorf("readed data while in wrong connection state %d", client.meshprotocol.connState))
			}
		} else {
			// No timeout if we don't have new data to send, send it now
			client.flushBuffer(client.inBuffer)
		}
	}

	client.socketWaitGroup.Done()
	if client.socketOpen {
		client.Close()
	}
}

func (client *OtaConnection) ForwardData(data []byte) error {
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

func NewOtaConnection(connection net.Conn, serial *SerialConnection, addr MeshNodeId, port int, closedCb func(NetworkConnection)) (*OtaConnection, error) {
	client := &OtaConnection{
		NetworkConnectionStruct: NewNetworkConnectionStruct(connection, serial, addr, port, closedCb),
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
