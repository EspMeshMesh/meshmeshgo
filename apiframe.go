package main

import (
	"encoding/binary"
	"errors"

	"github.com/go-restruct/restruct"
	"github.com/sirupsen/logrus"
)

type MeshNodeId uint32

type MeshProtocol byte

const (
	directProtocol MeshProtocol = iota
	bradcastProtocol
	unicastProtocol
	multipathProtocol
)

const startApiFrame byte = 0xFE
const escapeApiFrame byte = 0xEA
const stopApiFrame byte = 0xEF

const echoApiRequest uint8 = 0

type EchoApiRequest struct {
	Id   uint8  `struct:"uint8"`
	Echo string `struct:"string"`
}

const echoApiReply uint8 = 1

type EchoApiReply struct {
	Id   uint8  `struct:"uint8"`
	Echo string `struct:"string"`
}

const firmRevApiRequest uint8 = 2

type FirmRevApiRequest struct {
	Id uint8 `struct:"uint8"`
}

const firmRevApiReply uint8 = 3

type FirmRevApiReply struct {
	Id       uint8  `struct:"uint8"`
	Revision string `struct:"string"`
}

const nodeIdApiRequest uint8 = 4

type NodeIdApiRequest struct {
	Id uint8 `struct:"uint8"`
}

const nodeIdApiReply uint8 = 5

type NodeIdApiReply struct {
	Id     uint8      `struct:"uint8"`
	Serial MeshNodeId `struct:"uint32"`
}

const discoveryApiRequest uint8 = 26

type DiscoveryApiRequest struct {
	Id    uint8 `struct:"uint8"`
	ApiId uint8 `struct:"uint8"`
}

const discoveryApiReply uint8 = 27

type DiscoveryApiReply struct {
	Id    uint8 `struct:"uint8"`
	ApiId uint8 `struct:"uint8"`
}

const logEventApiReply uint8 = 57

type LogEventApiReply struct {
	Id    uint8      `struct:"uint8"`
	Level uint16     `struct:"uint16"`
	From  MeshNodeId `struct:"uint32"`
	Line  string     `struct:"string"`
}

const connectedUnicastRequest uint8 = 114

type UnicastRequest struct {
	Id      uint8      `struct:"uint8"`
	Target  MeshNodeId `struct:"uint32"`
	Payload []byte     `struct:"[]byte"`
}

//const connectedUnicastReply uint8 = 115

const multipathRequest uint8 = 118

type MultiPathRequest struct {
	Id      uint8      `struct:"uint8"`
	Target  MeshNodeId `struct:"uint32"`
	PathLen uint8      `struct:"uint8"`
	Path    []uint32   `struct:"[]uint32,sizefrom=[PathLen]"`
}

const meshmeshProtocolConnectedPath uint8 = 7

const connectedPathApiRequest uint8 = 122

type ConnectedPathApiRequest struct {
	Id       uint8  `struct:"uint8"`
	Protocol uint8  `struct:"uint8"`
	Command  uint8  `struct:"uint8"`
	Handle   uint16 `struct:"uint16"`
	Dummy    uint16 `struct:"uint16"`
	Sequence uint16 `struct:"uint16"`
	DataSize uint16 `struct:"uint16"`
	Data     []byte `struct:"[]byte,sizefrom=DataSize"`
}

type ConnectedPathApiRequest2 struct {
	Id       uint8   `struct:"uint8"`
	Protocol uint8   `struct:"uint8"`
	Command  uint8   `struct:"uint8"`
	Handle   uint16  `struct:"uint16"`
	Dummy    uint16  `struct:"uint16"`
	Sequence uint16  `struct:"uint16"`
	DataSize uint16  `struct:"uint16"`
	Port     uint16  `struct:"uint16"`
	PathLen  uint8   `struct:"uint8"`
	Path     []int32 `struct:"[]int32,sizefrom=PathLen"`
}

const connectedPathApiReply uint8 = 123

type ConnectedPathApiReply struct {
	Id      uint8  `struct:"uint8"`
	Command uint8  `struct:"uint8"`
	Handle  uint16 `struct:"uint16"`
	Data    []byte `struct:"[]byte"`
}

/* ----------------------------------------------------------------
   Discovery
 ---------------------------------------------------------------- */

const discResetTableApiRequest uint8 = 0x00

type DiscResetTableApiRequest struct {
	Id    uint8 `struct:"uint8"`
	ApiId uint8 `struct:"uint8"`
}

const discResetTableApiReply uint8 = 0x01

type DiscResetTableApiReply struct {
	Id    uint8 `struct:"uint8"`
	ApiId uint8 `struct:"uint8"`
}

const discTableSizeApiRequest uint8 = 0x02

type DiscTableSizeApiRequest struct {
	Id    uint8 `struct:"uint8"`
	ApiId uint8 `struct:"uint8"`
}

const discTableSizeApiReply uint8 = 0x03

type DiscTableSizeApiReply struct {
	Id    uint8 `struct:"uint8"`
	ApiId uint8 `struct:"uint8"`
	Size  uint8 `struct:"uint8"`
}

const discTableItemGetApiRequest uint8 = 0x04

type DiscTableItemGetApiRequest struct {
	Id    uint8 `struct:"uint8"`
	ApiId uint8 `struct:"uint8"`
	Index uint8 `struct:"uint8"`
}

const discTableItemGetApiReply uint8 = 0x05

type DiscTableItemGetApiReply struct {
	Id     uint8  `struct:"uint8"`
	ApiId  uint8  `struct:"uint8"`
	Index  uint8  `struct:"uint8"`
	NodeId uint32 `struct:"uint32"`
	Rssi1  int16  `struct:"int16"`
	Rssi2  int16  `struct:"int16"`
	Flags  uint16 `struct:"uint16"`
}

const discStartDiscoverApiRequest uint8 = 0x06

type DiscStartDiscoverApiRequest struct {
	Id      uint8 `struct:"uint8"`
	ApiId   uint8 `struct:"uint8"`
	Mask    uint8 `struct:"uint8"`
	Filter  uint8 `struct:"uint8"`
	Slotnum uint8 `struct:"uint8"`
}

const discStartDiscoverApiReply uint8 = 0x07

type DiscStartDiscoverApiReply struct {
	Id    uint8 `struct:"uint8"`
	ApiId uint8 `struct:"uint8"`
}

/* ----------------------------------------------------------------
   ApiFrame
 ---------------------------------------------------------------- */

type ApiFrame struct {
	data    []byte
	escaped bool
}

func (frame *ApiFrame) awaitedReplyBytes(index uint16) (uint8, uint8) {
	var wantType uint8 = frame.data[index]&0xFE + 1
	var wantSubtype uint8 = 0

	if wantSubtype == discoveryApiReply {
		wantSubtype = frame.data[index+1]&0xFE + 1
	}

	return wantType, wantSubtype
}

func (frame *ApiFrame) AwaitedReply() (uint8, uint8) {
	if len(frame.data) == 0 {
		return 0, 0
	} else {
		if frame.data[0] == connectedUnicastRequest {
			if len(frame.data) < 7 {
				return 0, 0
			} else {
				return frame.awaitedReplyBytes(6)
			}
		} else {
			return frame.awaitedReplyBytes(0)
		}
	}
}

func (frame *ApiFrame) AssertType(wantedType uint8, wantedSubtype uint8) bool {
	if len(frame.data) == 0 || frame.data[0] != wantedType && (wantedSubtype > 0 && (len(frame.data) < 2 || frame.data[1] != wantedSubtype)) {
		logrus.WithFields(logrus.Fields{"Want": wantedType, "Got": frame.data[0]}).Error("AssertType failed")
		return false
	} else {
		return true
	}
}

func (frame *ApiFrame) Escape() {
	if frame.escaped {
		return
	}

	var out []byte = []byte{}
	for _, b := range frame.data {
		if b == stopApiFrame || b == startApiFrame || b == escapeApiFrame {
			out = append(out, escapeApiFrame)
		}
		out = append(out, b)
	}

	frame.data = out
	frame.escaped = true
}

func (frame *ApiFrame) Output() []byte {
	if !frame.escaped {
		frame.Escape()
	}

	var out []byte = []byte{startApiFrame}
	out = append(out, frame.data...)
	out = append(out, stopApiFrame)
	return out
}

func (frame *ApiFrame) Decode() (interface{}, error) {
	if !frame.escaped {
		frame.Escape()
	}

	switch frame.data[0] {
	case echoApiReply:
		v := EchoApiReply{Id: 0, Echo: string(frame.data[1:])}
		return v, nil
	case firmRevApiReply:
		v := FirmRevApiReply{Id: 0, Revision: string(frame.data[1:])}
		return v, nil
	case nodeIdApiReply:
		v := NodeIdApiReply{}
		restruct.Unpack(frame.data, binary.LittleEndian, &v)
		return v, nil
	case logEventApiReply:
		v := LogEventApiReply{}
		restruct.Unpack(frame.data, binary.LittleEndian, &v)
		if len(frame.data) > 7 {
			v.Line = string(frame.data[7:])
		}
		return v, nil
	case connectedPathApiReply:
		v := ConnectedPathApiReply{}
		restruct.Unpack(frame.data, binary.LittleEndian, &v)
		if len(frame.data) > 4 {
			v.Data = frame.data[4:]
		}
		return v, nil
	case discoveryApiReply:
		v := DiscoveryApiReply{}
		restruct.Unpack(frame.data, binary.LittleEndian, &v)
		switch v.ApiId {
		case discResetTableApiReply:
			vv := DiscResetTableApiReply{}
			restruct.Unpack(frame.data, binary.LittleEndian, &vv)
			return vv, nil
		case discTableSizeApiReply:
			vv := DiscTableSizeApiReply{}
			restruct.Unpack(frame.data, binary.LittleEndian, &vv)
			return vv, nil
		case discTableItemGetApiReply:
			vv := DiscTableItemGetApiReply{}
			restruct.Unpack(frame.data, binary.LittleEndian, &vv)
			return vv, nil
		case discStartDiscoverApiReply:
			vv := DiscStartDiscoverApiReply{}
			restruct.Unpack(frame.data, binary.LittleEndian, &vv)
			return vv, nil
		}
	}

	return EchoApiReply{}, errors.New("unknow api frame")
}

func EncodeBuffer(cmd interface{}) ([]byte, error) {
	var b []byte
	var err error

	switch v := cmd.(type) {
	case EchoApiRequest:
		v.Id = echoApiRequest
		b, err = restruct.Pack(binary.LittleEndian, &v)
	case FirmRevApiRequest:
		v.Id = firmRevApiRequest
		b, err = restruct.Pack(binary.LittleEndian, &v)
	case NodeIdApiRequest:
		v.Id = nodeIdApiRequest
		b, err = restruct.Pack(binary.LittleEndian, &v)
	case ConnectedPathApiRequest:
		v.Id = connectedPathApiRequest
		b, err = restruct.Pack(binary.LittleEndian, &v)
	case ConnectedPathApiRequest2:
		v.Id = connectedPathApiRequest
		b, err = restruct.Pack(binary.LittleEndian, &v)
	case UnicastRequest:
		v.Id = connectedUnicastRequest
		b, err = restruct.Pack(binary.LittleEndian, &v)
	case MultiPathRequest:
		v.Id = multipathRequest
		b, err = restruct.Pack(binary.LittleEndian, &v)
	case DiscResetTableApiRequest:
		v.Id = discoveryApiRequest
		v.ApiId = discResetTableApiRequest
		b, err = restruct.Pack(binary.LittleEndian, &v)
	case DiscTableSizeApiRequest:
		v.Id = discoveryApiRequest
		v.ApiId = discTableSizeApiRequest
		b, err = restruct.Pack(binary.LittleEndian, &v)
	case DiscTableItemGetApiRequest:
		v.Id = discoveryApiRequest
		v.ApiId = discTableItemGetApiRequest
		b, err = restruct.Pack(binary.LittleEndian, &v)
	case DiscStartDiscoverApiRequest:
		v.Id = discoveryApiRequest
		v.ApiId = discStartDiscoverApiRequest
		b, err = restruct.Pack(binary.LittleEndian, &v)
	default:
		err = errors.New("unknow type request")
	}

	return b, err
}

func (frame *ApiFrame) EncodeFrame(cmd interface{}) error {
	b, err := EncodeBuffer(cmd)

	if err == nil {
		if len(b) == 0 {
			err = errors.New("can't encode requested stuct")
		} else {
			frame.data = b
			frame.escaped = true
		}
	}

	return err
}

func NewApiFrame(buffer []byte, escaped bool) *ApiFrame {
	f := &ApiFrame{
		data:    buffer,
		escaped: escaped,
	}

	return f
}

func NewApiFrameFromStruct(v interface{}, protocol MeshProtocol, target MeshNodeId) (*ApiFrame, error) {
	var err error
	f := &ApiFrame{}
	if protocol == directProtocol {
		err = f.EncodeFrame(v)
	} else if protocol == unicastProtocol {
		p := UnicastRequest{Id: connectedUnicastRequest, Target: target}
		p.Payload, err = EncodeBuffer(v)
		if err == nil {
			err = f.EncodeFrame(p)
		}
	} else {
		log.Error("Unknow protocol requested")
	}
	return f, err
}
