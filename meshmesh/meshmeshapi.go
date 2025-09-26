package meshmesh

import (
	"container/list"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-restruct/restruct"
	"github.com/sirupsen/logrus"
	"go.bug.st/serial"
	"leguru.net/m/v2/graph"
	"leguru.net/m/v2/logger"
)

const defaultSessionMaxTimeoutMs = 500
const maxSerialInputBuffer = 8192

type SerialSession struct {
	Request      *ApiFrame
	Reply        *ApiFrame
	WaitReply1   uint8
	WaitReply2   uint8
	Wait         sync.WaitGroup
	SentTime     time.Time
	MaxTimeoutMs int64
}

func (session *SerialSession) IsAwaitable() bool {
	return session.WaitReply1 > 0
}

func NewSimpleSerialSession(request *ApiFrame) *SerialSession {
	s := SerialSession{Request: request, MaxTimeoutMs: defaultSessionMaxTimeoutMs}
	return &s
}

func NewSerialSession(request *ApiFrame) (*SerialSession, error) {
	w1, w2, err := request.AwaitedReply()
	if err != nil {
		return nil, err
	}
	s := SerialSession{Request: request, WaitReply1: w1, WaitReply2: w2}
	s.MaxTimeoutMs = defaultSessionMaxTimeoutMs
	s.SentTime = time.Now()
	return &s, nil
}

type SerialConnection struct {
	connected       bool
	port            serial.Port
	isEsp8266       bool
	txOneByteMs     int
	debug           bool
	incoming        chan []byte
	session         *SerialSession
	Sessions        *list.List
	SessionsLock    sync.Mutex
	NextHandle      uint16
	LocalNode       uint32
	ConnPathFn      func(*ConnectedPathApiReply)
	DiscAssociateFn func(*DiscAssociateApiReply, *SerialConnection)
}

const (
	waitStartByte = iota
	escapeNextByte
	waitEndByte
	waitCrc16Byte1
	waitCrc16Byte2
	waitEndOfLine
)

func (serialConn *SerialConnection) IsConnected() bool {
	return serialConn.connected
}

func (serialConn *SerialConnection) GetNextHandle() uint16 {
	nh := serialConn.NextHandle
	serialConn.NextHandle += 1
	// Never use handle 0
	if serialConn.NextHandle == 0 {
		serialConn.NextHandle += 1
	}
	return nh
}

func (serialConn *SerialConnection) ReadFrame(buffer []byte) {
	frame := NewApiFrame(buffer, true)
	if buffer[0] != logEventApiReply {
		logger.Log().WithFields(logrus.Fields{"len": len(frame.data), "data": hex.EncodeToString(frame.data[0:min(len(frame.data), 10)])}).Trace("From serial")
	}
	switch buffer[0] {
	case logEventApiReply:
		// Handle LOG packets first
		v, err := frame.Decode()
		if err != nil {
			logger.Log().Error("Can't decode incoming log packet 1/2")
		} else {
			lo, ok := v.(LogEventApiReply)
			if !ok {
				logger.Log().Error("Can't decode incoming log packet 2/2")
			}
			logger.Log().WithFields(logrus.Fields{"from": lo.From}).Debug(lo.Line)
		}
	case connectedPathApiReply:
		// Handle ConnectedPath packets next
		v, err := frame.Decode()
		if err != nil {
			logger.Log().Error("Can't decode incoming connectedpath packet 1/2")
		} else {
			c, ok := v.(ConnectedPathApiReply)
			if !ok {
				logger.Log().Error("Can't decode incoming connectedpath packet 2/2")
			}
			if serialConn.ConnPathFn != nil {
				serialConn.ConnPathFn(&c)
			}
		}
	default:
		// Handle session pacekts next
		if serialConn.session != nil {
			if serialConn.session.WaitReply1 > 0 {
				if frame.AssertType(serialConn.session.WaitReply1, serialConn.session.WaitReply2) {
					serialConn.session.Reply = frame
					serialConn.session.Wait.Done()
					serialConn.session = nil
				} else {
					logger.Log().WithFields(logrus.Fields{"Type": serialConn.session.WaitReply1, "Subtype": serialConn.session.WaitReply2}).Error("Serial reply assertion failed")
				}
			}
		} else {
			if frame.AssertType(discoveryApiReply, discResetTableApiReply) {
				vv := DiscAssociateApiReply{}
				restruct.Unpack(frame.data, binary.LittleEndian, &vv)
				if serialConn.DiscAssociateFn != nil {
					serialConn.DiscAssociateFn(&vv, serialConn)
				}
			} else {
				logger.Log().WithField("type", fmt.Sprintf("%02X", buffer[0])).Error("Unused packet received")
			}
		}
	}
}

func (conn *SerialConnection) checkSessionTimeout() {
	if conn.session != nil {
		if time.Since(conn.session.SentTime).Milliseconds() > conn.session.MaxTimeoutMs {
			conn.session.Reply = nil
			if conn.session.WaitReply1 > 0 {
				conn.session.Wait.Done()
			}
			conn.session = nil
		}
	}
}

func (serialConn *SerialConnection) processSerialBuffer(buffer []byte, bufferPos int) {
	destination := make([]byte, bufferPos)
	copy(destination, buffer)
	serialConn.ReadFrame(destination)
}

func (serialConn *SerialConnection) Read() {
	var lastStartByte uint8
	var computedCrc16 uint16
	var receivedCrc16 uint16
	var inputBufferPos int
	inputBuffer := make([]byte, maxSerialInputBuffer)
	var decodeState int = waitStartByte
	serialConn.port.ResetInputBuffer()

	for {
		var buffer = make([]byte, 1)
		// Read a byte from serial with a timout of a time slot
		serialConn.port.SetReadTimeout(50 * time.Millisecond)
		n, err := serialConn.port.Read(buffer)
		if err != nil {
			break
		}

		if n == 0 {
			// We don't receive any data check if we want a reply
			serialConn.checkSessionTimeout()
		} else if n > 0 {
			b := buffer[0]
			switch decodeState {
			case waitStartByte:
				switch b {
				case startApiFrame:
					lastStartByte = b
					inputBufferPos = 0
					computedCrc16 = 0
					decodeState = waitEndByte
				case startApiFrameCrc16:
					lastStartByte = b
					inputBufferPos = 0
					computedCrc16 = 0
					decodeState = waitEndByte
				case startLogMsg:
					lastStartByte = b
					inputBufferPos = 0
					computedCrc16 = 0
					decodeState = waitEndOfLine
					inputBuffer[inputBufferPos] = b
					inputBufferPos += 1

				default:
					logger.Log().WithField("b", fmt.Sprintf("0x%02X ", b)).Error("serial error: received a character outside a frame")
				}
			case escapeNextByte:
				decodeState = waitEndByte
				// And escaped byte is take as is not used for commands.
				inputBuffer[inputBufferPos] = b
				computedCrc16 = crc16Byte(computedCrc16, b)
				inputBufferPos += 1
			case waitCrc16Byte1:
				receivedCrc16 = uint16(b) << 8
				decodeState = waitCrc16Byte2
			case waitCrc16Byte2:
				receivedCrc16 = receivedCrc16 | uint16(b)
				if receivedCrc16 == computedCrc16 {
					decodeState = waitStartByte
					serialConn.processSerialBuffer(inputBuffer, inputBufferPos)
					inputBufferPos = 0
				} else {
					decodeState = waitStartByte
					logger.Log().WithFields(logrus.Fields{"receivedCrc16": receivedCrc16, "computedCrc16": computedCrc16}).Error("serial error: crc16 mismatch")
					inputBufferPos = 0
				}
			case waitEndByte:
				switch b {
				case stopApiFrame:
					if lastStartByte == startApiFrameCrc16 {
						// Wait for two more bytes to complete the crc16
						decodeState = waitCrc16Byte1
					} else {
						// No crc16, just process the buffer
						serialConn.processSerialBuffer(inputBuffer, inputBufferPos)
						inputBufferPos = 0
						decodeState = waitStartByte
					}
				case escapeApiFrame:
					decodeState = escapeNextByte
					computedCrc16 = crc16Byte(computedCrc16, b)
				default:
					inputBuffer[inputBufferPos] = b
					inputBufferPos += 1
					computedCrc16 = crc16Byte(computedCrc16, b)
				}
			case waitEndOfLine:
				if b == stopLogMsg {
					destination := make([]byte, inputBufferPos)
					copy(destination, inputBuffer)
					fmt.Println("==> " + string(destination))
					decodeState = waitStartByte
				} else {
					inputBuffer[inputBufferPos] = b
					inputBufferPos += 1
				}

			default:
				switch b {
				case stopApiFrame:
					if lastStartByte == startApiFrameCrc16 {
						// Wait for two more bytes to complete the crc16
						decodeState = waitCrc16Byte1
					} else {
						// No crc16, just process the buffer
						serialConn.processSerialBuffer(inputBuffer, inputBufferPos)
						inputBufferPos = 0
						decodeState = waitStartByte
					}
				case escapeApiFrame:
					decodeState = escapeNextByte
					computedCrc16 = crc16Byte(computedCrc16, b)
				default:
					inputBuffer[inputBufferPos] = b
					inputBufferPos += 1
					computedCrc16 = crc16Byte(computedCrc16, b)
				}
			}

			if inputBufferPos >= 1500 {
				logger.Log().WithFields(logrus.Fields{"buffer": hex.EncodeToString(inputBuffer)}).Error("Buffer overflow")
				decodeState = waitStartByte
				inputBufferPos = 0
			}
		}
	}

	serialConn.connected = false
	serialConn.port.Close()
	logger.Log().Info("SerialConnection.Read terminated")
}

func (serialConn *SerialConnection) Write() {
	for {
		// If we are idle
		if serialConn.session == nil {
			// And there is not more work to do
			if serialConn.Sessions.Len() == 0 {
				// Sleep a time slot
				time.Sleep(50 * time.Millisecond)
			} else {
				// We are idle but we have work to do...
				serialConn.SessionsLock.Lock()
				element := serialConn.Sessions.Front().Value
				// Remove from sessions list
				serialConn.Sessions.Remove(serialConn.Sessions.Front())
				serialConn.SessionsLock.Unlock()

				if element == nil {
					// Ok we don't really need this
					logger.Log().WithFields(logrus.Fields{"queue": serialConn.Sessions.Len()}).Error("got session with nil value")
					// Sleep a time slot
					time.Sleep(50 * time.Millisecond)
				} else {
					// Get next session and remove from list
					session, ok := element.(*SerialSession)

					if ok {
						b := session.Request.Output()
						level := logger.Log().GetLevel()
						if level >= logrus.TraceLevel {
							logger.Log().WithFields(logrus.Fields{"len": len(b), "data": hex.EncodeToString(b[0:min(len(b), 32)])}).Trace("To serial")
						}

						writed, err := serialConn.port.Write(b)

						if err != nil {
							logger.Log().WithField("err", err).Error("Write to serial port error")
							break
						}

						if writed < len(b) {
							logger.Log().WithFields(logrus.Fields{"sent": writed, "want": len(b)}).Error("Write to serial port incomplete")
							break
						}

						if session.WaitReply1 > 0 {
							// If we need a reply mark we as busy
							serialConn.session = session
						} else {
							// Sleep a time slot beofre send next session
							// Is a guard time for wifi retransmissions
							time.Sleep(50 * time.Millisecond)
						}
					} else {
						// Ok we don't really need this
						logger.Log().WithFields(logrus.Fields{"queue": serialConn.Sessions.Len(), "val": element}).Error("interface conversion invalid")
						// Sleep a time slot
						time.Sleep(50 * time.Millisecond)
					}
				}

			}
		} else {
			// We are busy Sleep a time slot before check again
			time.Sleep(50 * time.Millisecond)
		}
	}

	serialConn.connected = false
	serialConn.port.Close()
	logger.Log().Info("SerialConnection.Write terminated")
}

func (serialConn *SerialConnection) QueueApiSession(session *SerialSession) {
	serialConn.SessionsLock.Lock()
	serialConn.Sessions.PushBack(session)
	serialConn.SessionsLock.Unlock()
}

func (serialConn *SerialConnection) SendApi(cmd interface{}) error {
	frame, err := NewApiFrameFromStruct(cmd, DirectProtocol, 0, nil)
	if err != nil {
		return err
	}

	session := NewSimpleSerialSession(frame)
	serialConn.QueueApiSession(session)
	return nil
}

func (serialConn *SerialConnection) sendReceiveApiProt(session *SerialSession) (interface{}, error) {
	if session.IsAwaitable() {
		session.Wait.Add(1)
	}
	serialConn.QueueApiSession(session)
	if session.IsAwaitable() {
		session.Wait.Wait()
	}

	if session.Reply == nil {
		return nil, errors.New("reply timeout")
	} else {
		return session.Reply.Decode()
	}
}

func (serialConn *SerialConnection) SendReceiveApiProt(cmd any, protocol MeshProtocol, target MeshNodeId, network *graph.Network) (interface{}, error) {
	if target == 0 {
		protocol = DirectProtocol
	}
	frame, err := NewApiFrameFromStruct(cmd, protocol, target, network)
	if err != nil {
		return nil, err
	}

	session, err := NewSerialSession(frame)
	if err != nil {
		return nil, err
	}

	return serialConn.sendReceiveApiProt(session)
}

func (serialConn *SerialConnection) SendReceiveApiProtTimeout(cmd interface{}, protocol MeshProtocol, target MeshNodeId, network *graph.Network, timeoutMs int64) (interface{}, error) {
	if target == 0 {
		protocol = DirectProtocol
	}
	frame, err := NewApiFrameFromStruct(cmd, protocol, target, network)
	if err != nil {
		return nil, err
	}

	session, err := NewSerialSession(frame)
	if err != nil {
		return nil, err
	}

	session.MaxTimeoutMs = timeoutMs
	return serialConn.sendReceiveApiProt(session)
}

func (serialConn *SerialConnection) SendReceiveApi(cmd interface{}) (interface{}, error) {
	return serialConn.SendReceiveApiProt(cmd, DirectProtocol, 0, nil)
}

func NewSerial(portName string, baudRate int, isEsp8266 bool, debug bool) (*SerialConnection, error) {
	mode := &serial.Mode{BaudRate: baudRate}
	p, err := serial.Open(portName, mode)
	if err != nil {
		return nil, err
	}

	serial := &SerialConnection{
		connected:   true,
		port:        p,
		isEsp8266:   isEsp8266,
		txOneByteMs: int(float32(8) / float32(baudRate) * 1000000.0),
		debug:       debug,
		incoming:    make(chan []byte),
		Sessions:    list.New(),
		NextHandle:  1,
	}

	go serial.Write()
	go serial.Read()

	reply1, err := serial.SendReceiveApi(EchoApiRequest{Echo: "CIAO"})
	if err != nil {
		serial.port.Close()
		return nil, err
	}
	echo, ok := reply1.(EchoApiReply)
	if !ok {
		serial.port.Close()
		return nil, errors.New("invalid echo reply type")
	}
	if echo.Echo != "CIAO" {
		serial.port.Close()
		return nil, errors.New("invalid echo reply")
	}

	reply2, err := serial.SendReceiveApi(NodeIdApiRequest{})
	if err != nil {
		serial.port.Close()
		return nil, err
	}

	nodeid, ok := reply2.(NodeIdApiReply)
	if !ok {
		serial.port.Close()
		return nil, errors.New("invalid nodeid reply")
	}

	reply3, err := serial.SendReceiveApi(FirmRevApiRequest{})
	if err != nil {
		serial.port.Close()
		return nil, err
	}
	firmrev, ok := reply3.(FirmRevApiReply)
	if !ok {
		serial.port.Close()
		return nil, errors.New("invalid firmware reply")
	}

	serial.LocalNode = uint32(nodeid.Serial)
	logger.Log().WithFields(logrus.Fields{"nodeId": fmt.Sprintf("0x%06X", serial.LocalNode), "firmware": firmrev.Revision}).
		Info("Valid local node found")
	return serial, nil
}
