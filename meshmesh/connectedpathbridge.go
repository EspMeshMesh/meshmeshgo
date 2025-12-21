package meshmesh

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"leguru.net/m/v2/graph"
	"leguru.net/m/v2/logger"
	"leguru.net/m/v2/utils"
)

type ConnectionPathBridgeDriver interface {
	Socket2Serial(buffer *bytes.Buffer, connectedPath *ConnPathConnection, stats *EspApiConnectionStats)
}

type ConnectionPathBridge struct {
	Stats           *EspApiConnectionStats
	tmpBuffer       *bytes.Buffer
	inBuffer        *bytes.Buffer
	socketOpen      bool
	socket          net.Conn
	socketWaitGroup sync.WaitGroup
	connectedPath   *ConnPathConnection
	reqAddress      MeshNodeId
	reqPort         int
	debugThisNode   bool
	timeout         time.Time
	clientClosed    func(client *ConnectionPathBridge)
	driver          ConnectionPathBridgeDriver
}

func (c *ConnectionPathBridge) serialDataAvailable(data []byte) {
	logger.WithFields(logger.Fields{
		"handle": c.connectedPath.handle,
		"meshid": utils.FmtNodeId(int64(c.reqAddress)),
		"len":    len(data),
		"data":   utils.EncodeToHexEllipsis(data, 10),
	}).Trace("SE-->HA")
	n, err := c.socket.Write(data)
	if err != nil || n < len(data) {
		logger.WithFields(logger.Fields{"handle": c.connectedPath.handle, "err": err}).Error("serialDataAvailable error")
	}
}

func (c *ConnectionPathBridge) close() {
	c.socketOpen = false
	c.socket.Close()
	c.socketWaitGroup.Wait()
	c.Stats.Stop()
	c.connectedPath.Disconnect()
	c.clientClosed(c)
	logger.Log().Debug("ConnectionPathBridgeBase.close")
}

func (c *ConnectionPathBridge) connectionActive() {
	c.finishHandshake(true)
}

func (c *ConnectionPathBridge) connectionInvalid() {
	c.close()
}

func (c *ConnectionPathBridge) Dispose() {
	c.connectedPath.Dispose()
}

func (c *ConnectionPathBridge) startHandshake(addr MeshNodeId, port int) error {
	c.reqAddress = addr
	c.reqPort = port
	err := c.connectedPath.OpenConnectionAsync(addr, uint16(port))
	if err == nil {
		c.Stats.Start()
		if addr == MeshNodeId(0) {
			c.debugThisNode = true
			logger.WithFields(logger.Fields{"id": fmt.Sprintf("%02X", addr)}).Info("startHandshake and debug for node")
		}
	}
	return err
}

func (c *ConnectionPathBridge) finishHandshake(result bool) {
	logger.WithField("res", result).Debug("finishHandshake")
	if !result {
		logger.WithFields(logger.Fields{"addr": c.reqAddress, "port": c.reqPort, "err": nil}).
			Warning("ApiConnection.finishHandshake failed")
	} else {
		logger.WithFields(logger.Fields{"addr": c.reqAddress, "port": c.reqPort, "handle": c.connectedPath.handle}).
			Info("ApiConnection.handshake OpenConnection succesfull")
		c.driver.Socket2Serial(c.tmpBuffer, c.connectedPath, c.Stats)
		c.Stats.GotHandle(c.connectedPath.handle)
	}
}

// Go routine to read data from socket and forward it to the connected path serial protcol
// This function read form scoket and forward it to the connected path serial protcol
// If the connected path is in handshake state, the data is stored in a temporary buffer
// The temporary buffer is sent when the handshake is finished @see finishHandshake
func (c *ConnectionPathBridge) readFromSocketRoutine() {
	var n int
	var err error

	for {
		var buffer = make([]byte, 1)
		c.socket.SetReadDeadline(time.Now().Add(10 * time.Millisecond))

		n, err = c.socket.Read(buffer)
		c.Stats.ReceivedBytes(n)

		if err == io.EOF {
			logger.WithFields(logger.Fields{"handle": c.socket.RemoteAddr().String()}).Warn("ConnectionPathBridgeBase.Read connection closed by peer")
			break
		}

		if err != nil {
			if !errors.Is(err, os.ErrDeadlineExceeded) {
				logger.WithFields(logger.Fields{"handle": c.socket.RemoteAddr().String(), "err": err}).Error("ConnectionPathBridgeBase.Read error")
				break
			}
		}

		if n > 0 {
			switch c.connectedPath.connState {
			case connPathConnectionStateHandshakeStarted:
				// FIXME check for if buffer grown outside limits
				c.tmpBuffer.WriteByte(buffer[0])
			case connPathConnectionStateActive:
				// FIXME handle error
				c.inBuffer.WriteByte(buffer[0])
			default:
				logger.WithField("state", c.connectedPath.connState).
					Error(fmt.Errorf("readed data while in wrong connection state %d", c.connectedPath.connState))
			}
		} else {
			if c.connectedPath.connState == connPathConnectionStateActive {
				// timeout reached if we have new data to send, send it now
				// tmpBuffer is sent when handshake is finished
				if c.inBuffer.Len() > 0 {
					c.driver.Socket2Serial(c.inBuffer, c.connectedPath, c.Stats)
				}
			}
		}
	}

	c.socketWaitGroup.Done()
	if c.socketOpen {
		c.close()
	}
}

// Go routine to check if the connection has timed out
func (c *ConnectionPathBridge) checkTimeoutRoutine() {
	for {
		if !c.socketOpen {
			break
		}
		if c.connectedPath.connState == connPathConnectionStateInit || c.connectedPath.connState == connPathConnectionStateHandshakeStarted {
			if time.Since(c.timeout).Milliseconds() > 3000 {
				logger.Error(fmt.Sprintf("Closing connection beacuse timeout after %dms in connPathConnectionStateInit for handle %d", time.Since(c.timeout).Milliseconds(), c.connectedPath.handle))
				c.close()
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	logger.Debug("ApiConnection.CheckTimeout exited")
}

func NewConnectionPathBridge(socket net.Conn, serialProxy *ConnectedPath2Serial, network *graph.Network, addr MeshNodeId, port int, driver ConnectionPathBridgeDriver, closedCb func(*ConnectionPathBridge)) (*ConnectionPathBridge, error) {
	if !serialProxy.IsSerialConnected() {
		return nil, errors.New("serial is not open")
	}

	conn := ConnectionPathBridge{
		connectedPath: NewConnPathConnection(serialProxy, network),
		socketOpen:    true,
		socket:        socket,
		tmpBuffer:     bytes.NewBuffer([]byte{}),
		inBuffer:      bytes.NewBuffer([]byte{}),
		timeout:       time.Now(),
		clientClosed:  closedCb,
		Stats:         _allStats.Stats(addr),
		driver:        driver,
	}

	conn.connectedPath.SetSerialDataAvailableCallback(conn.serialDataAvailable)
	conn.connectedPath.SetConnectionActiveCallback(conn.connectionActive)
	conn.connectedPath.SetConnectionInvalidCallback(conn.connectionInvalid)

	err := conn.startHandshake(addr, port)
	if err != nil {
		return nil, err
	}

	conn.socketWaitGroup.Add(1)
	go conn.readFromSocketRoutine()
	go conn.checkTimeoutRoutine()

	return &conn, nil
}
