package meshmesh

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/go-restruct/restruct"
	"google.golang.org/protobuf/proto"
	"leguru.net/m/v2/graph"
	pb "leguru.net/m/v2/meshmesh/pb"
)

type MeshNodeId uint32

type MeshProtocol byte

const (
	AutoProtocol MeshProtocol = iota
	DirectProtocol
	BradcastProtocol
	UnicastProtocol
	MultipathProtocol
)

const (
	startApiFrameCrc16 byte = 0xFD
	startApiFrame      byte = 0xFE
	escapeApiFrame     byte = 0xEA
	stopApiFrame       byte = 0xEF
	startLogMsg        byte = 0x1B
	stopLogMsg         byte = 0x0A
)

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

const nodeGetTagApiRequest uint8 = 6

type NodeGetTagApiRequest struct {
	Id uint8 `struct:"uint8"`
}

const nodeGetTagApiReply = 7

type NodeGetTagApiReply struct {
	Id  uint8  `struct:"uint8"`
	Tag []byte `struct:"[31]byte"`
}

const nodeSetTagApiRequest = 8

type NodeSetTagApiRequest struct {
	Id  uint8  `struct:"uint8"`
	Tag string `struct:"[31]byte"`
}

const nodeSetTagApiReply = 9

type NodeSetTagApiReply struct {
	Id uint8 `struct:"uint8"`
}

const nodeBindClearApiRequest = 10

type NodeBindClearApiRequest struct {
	Id uint8 `struct:"uint8"`
}

const nodeBindClearApiReply = 11

type NodeBindClearApiReply struct {
	Id uint8 `struct:"uint8"`
}

const nodeSetChannelApiRequest = 12

type NodeSetChannelApiRequest struct {
	Id      uint8 `struct:"uint8"`
	Channel uint8 `struct:"int8"`
}

const nodeSetChannelApiReply = 13

type NodeSetChannelApiReply struct {
	Id uint8 `struct:"uint8"`
}

const nodeConfigApiRequest uint8 = 14

type NodeConfigApiRequest struct {
	Id uint8 `struct:"uint8"`
}

const nodeConfigApiReply uint8 = 15

type NodeConfigApiReply struct {
	Id           uint8  `struct:"uint8"`
	Tag          []byte `struct:"[32]byte"`
	LogDest      uint32 `struct:"uint32"`
	Channel      uint8  `struct:"uint8"`
	TxPower      uint8  `struct:"uint8"`
	Groups       uint32 `struct:"uint32"`
	BindedServer uint32 `struct:"uint32"`
	Flags        uint8  `struct:"uint8"`
}

const nodeRebootApiRequest uint8 = 24

type NodeRebootApiRequest struct {
	Id uint8 `struct:"uint8"`
}

const nodeRebootApiReply uint8 = 25

type NodeRebootApiReply struct {
	Id uint8 `struct:"uint8"`
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

const flashOperationApiRequest uint8 = 30

const flashOperationApiReply uint8 = 31

type FlashOperationApiReply struct {
	Id    uint8 `struct:"uint8"`
	ApiId uint8 `struct:"uint8"`
}

const entitiesCountApiRequest uint8 = 38

type EntitiesCountApiRequest struct {
	Id uint8 `struct:"uint8"`
}

const entitiesCountApiReply uint8 = 39

type EntitiesCountApiReply struct {
	Id       uint8   `struct:"uint8"`
	Counters []uint8 `struct:"[6]uint8"`
}

const entityHashApiRequest uint8 = 40

type EntityHashApiRequest struct {
	Id      uint8 `struct:"uint8"`
	Service uint8 `struct:"uint8"`
	Index   uint8 `struct:"uint8"`
}

const entityHashApiReply uint8 = 41

type EntityHashApiReply struct {
	Id   uint8  `struct:"uint8"`
	Hash uint16 `struct:"uint16"`
	Info string `struct:"string"`
}

const getEntityStateApiRequest uint8 = 42

type GetEntityStateApiRequest struct {
	Id      uint8  `struct:"uint8"`
	Service uint8  `struct:"uint8"`
	Hash    uint16 `struct:"uint16"`
}

const getEntityStateApiReply uint8 = 43

type GetEntityStateApiReply struct {
	Id    uint8  `struct:"uint8"`
	State uint16 `struct:"uint16"`
}

const setEntityStateApiRequest uint8 = 44

type SetEntityStateApiRequest struct {
	Id      uint8  `struct:"uint8"`
	Service uint8  `struct:"uint8"`
	Hash    uint16 `struct:"uint16"`
	State   uint16 `struct:"uint16"`
}

const setEntityStateApiReply uint8 = 45

type SetEntityStateApiReply struct {
	Id uint8 `struct:"uint8"`
}

const logEventApiReply uint8 = 57

type LogEventApiReply struct {
	Id    uint8      `struct:"uint8"`
	Level uint16     `struct:"uint16"`
	From  MeshNodeId `struct:"uint32"`
	Line  string     `struct:"string"`
}

const nodePresentationApiReply uint8 = 65

const protoPresentationRxApiReply uint8 = 69

type NodePresentationApiReply struct {
	Id          uint8          `struct:"uint8"`
	SourceAddr  MeshNodeId     `struct:"uint32"`
	TargetAddr  MeshNodeId     `struct:"uint32"`
	Repeaters   [16]MeshNodeId `struct:"[16]uint32"`
	Rssi        [16]int16      `struct:"[16]int16"`
	Hops        uint8          `struct:"uint8"`
	Hostname    [16]byte       `struct:"[16]byte"`
	FwVersion   [16]byte       `struct:"[16]byte"`
	CompileTime [24]byte       `struct:"[24]byte"`
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
	Path    []uint32   `struct:"[]uint32,sizefrom=PathLen"`
	Payload []byte     `struct:"[]byte"`
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

const discAssociateApiReply uint8 = 0x0B

type DiscAssociateApiReply struct {
	Id     uint8         `struct:"uint8"`
	ApiId  uint8         `struct:"uint8"`
	Source MeshNodeId    `struct:"uint32"`
	Server MeshNodeId    `struct:"uint32"`
	Rssi   [3]int16      `struct:"[3]int16"`
	NodeId [3]MeshNodeId `struct:"[3]uint32"`
}

/* ----------------------------------------------------------------
   Flash
 ---------------------------------------------------------------- */

const flashGetMd5Api uint8 = 1

type FlashGetMd5ApiRequest struct {
	Id      uint8  `struct:"uint8"`
	ApiId   uint8  `struct:"uint8"`
	Address uint32 `struct:"uint32"`
	Length  uint32 `struct:"uint32"`
}

type FlashGetMd5ApiReply struct {
	Id     uint8  `struct:"uint8"`
	ApiId  uint8  `struct:"uint8"`
	Erased bool   `struct:"bool"`
	MD5    []byte `struct:"[16]byte"`
}

const flashEraseApi uint8 = 2

type FlashEraseApiRequest struct {
	Id      uint8  `struct:"uint8"`
	ApiId   uint8  `struct:"uint8"`
	Address uint32 `struct:"uint32"`
	Length  uint32 `struct:"uint32"`
}

type FlashEraseApiReply struct {
	Id     uint8 `struct:"uint8"`
	ApiId  uint8 `struct:"uint8"`
	Erased uint8 `struct:"uint8"`
}

const flashWriteApi uint8 = 3

type FlashWriteApiRequest struct {
	Id      uint8  `struct:"uint8"`
	ApiId   uint8  `struct:"uint8"`
	Address uint32 `struct:"uint32"`
	Data    []byte `struct:"[]byte"`
}

type FlashWriteApiReply struct {
	Id     uint8 `struct:"uint8"`
	ApiId  uint8 `struct:"uint8"`
	Result bool  `struct:"bool"`
}

const flashEBootApiRequest uint8 = 4

type FlashEBootApiRequest struct {
	Id      uint8  `struct:"uint8"`
	ApiId   uint8  `struct:"uint8"`
	Address uint32 `struct:"uint32"`
	Length  uint32 `struct:"uint32"`
}

type FlashEBootApiReply struct {
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

func (frame *ApiFrame) awaitedReplyBytes(index uint16) (uint8, uint8, error) {
	var wantType uint8 = frame.data[index]&0xFE + 1
	var wantSubtype uint8 = 0

	switch wantType {
	case discoveryApiReply:
		wantSubtype = frame.data[index+1]&0xFE + 1
	case flashOperationApiReply:
		wantSubtype = frame.data[index+1]&0xFE + 1
	}

	return wantType, wantSubtype, nil
}

func (frame *ApiFrame) AwaitedReply() (uint8, uint8, error) {
	if len(frame.data) == 0 {
		return 0, 0, errors.New("can't send an empty frame")
	} else {
		if frame.data[0] == connectedUnicastRequest {
			if len(frame.data) < 6 {
				return 0, 0, errors.New("invalid unicast frame")
			} else {
				return frame.awaitedReplyBytes(5)
			}
		} else {
			return frame.awaitedReplyBytes(0)
		}
	}
}
func (frame *ApiFrame) AssertType(wantedType uint8, wantedSubtype uint8) bool {
	if len(frame.data) == 0 || (frame.data[0] != wantedType) || (wantedSubtype > 0 && (len(frame.data) < 2 || frame.data[1] != wantedSubtype)) {
		return false
	} else {
		return true
	}
}

func (frame *ApiFrame) Escape() {
	if frame.escaped {
		return
	}

	var escapes = 0
	for _, b := range frame.data {
		if b == stopApiFrame || b == startApiFrame || b == escapeApiFrame {
			escapes += 1
		}
	}

	var j = 0
	escaped := make([]byte, len(frame.data)+escapes)
	for _, b := range frame.data {
		if b == stopApiFrame || b == startApiFrame || b == escapeApiFrame {
			escaped[j] = escapeApiFrame
			j += 1
		}
		escaped[j] = b
		j += 1
	}

	frame.data = escaped
	frame.escaped = true
}

func (frame *ApiFrame) Output() []byte {
	if !frame.escaped {
		frame.Escape()
	}

	crc16hash := crc16Calc(0, frame.data, len(frame.data))

	var out []byte = []byte{startApiFrameCrc16}
	out = append(out, frame.data...)
	out = append(out, stopApiFrame, byte(crc16hash>>8), byte(crc16hash))
	return out
}

func (frame *ApiFrame) Decode() (any, error) {
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
	case nodeGetTagApiReply:
		v := NodeGetTagApiReply{}
		restruct.Unpack(frame.data, binary.LittleEndian, &v)
		return v, nil
	case nodeSetTagApiReply:
		v := NodeSetTagApiReply{}
		restruct.Unpack(frame.data, binary.LittleEndian, &v)
		return v, nil
	case nodeBindClearApiReply:
		v := NodeBindClearApiReply{}
		restruct.Unpack(frame.data, binary.LittleEndian, &v)
		return v, nil
	case nodeSetChannelApiReply:
		v := NodeSetChannelApiReply{}
		restruct.Unpack(frame.data, binary.LittleEndian, &v)
		return v, nil
	case nodeConfigApiReply:
		v := NodeConfigApiReply{}
		restruct.Unpack(frame.data, binary.LittleEndian, &v)
		return v, nil
	case nodeRebootApiReply:
		v := NodeRebootApiReply{}
		restruct.Unpack(frame.data, binary.LittleEndian, &v)
		return v, nil
	case entitiesCountApiReply:
		v := EntitiesCountApiReply{}
		restruct.Unpack(frame.data, binary.LittleEndian, &v)
		return v, nil
	case entityHashApiReply:
		v := EntityHashApiReply{}
		restruct.Unpack(frame.data, binary.LittleEndian, &v)
		return v, nil
	case getEntityStateApiReply:
		v := GetEntityStateApiReply{}
		restruct.Unpack(frame.data, binary.LittleEndian, &v)
		return v, nil
	case setEntityStateApiReply:
		v := SetEntityStateApiReply{}
		restruct.Unpack(frame.data, binary.LittleEndian, &v)
		return v, nil
	case logEventApiReply:
		v := LogEventApiReply{}
		restruct.Unpack(frame.data, binary.LittleEndian, &v)
		if len(frame.data) > 7 {
			v.Line = string(frame.data[7:])
		}
		return v, nil
	case nodePresentationApiReply:
		v := NodePresentationApiReply{}
		restruct.Unpack(frame.data, binary.LittleEndian, &v)
		return v, nil
	case protoPresentationRxApiReply:
		v := pb.NodePresentationRx{}
		err := proto.Unmarshal(frame.data, &v)
		if err != nil {
			return nil, err
		}
		return &v, nil
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
		case discAssociateApiReply:
			vv := DiscAssociateApiReply{}
			restruct.Unpack(frame.data, binary.LittleEndian, &vv)
			return vv, nil
		}
	case flashOperationApiReply:
		v := FlashOperationApiReply{}
		restruct.Unpack(frame.data, binary.LittleEndian, &v)
		switch v.ApiId {
		case flashGetMd5Api:
			vv := FlashGetMd5ApiReply{}
			restruct.Unpack(frame.data, binary.LittleEndian, &vv)
			return vv, nil
		case flashEraseApi:
			vv := FlashEraseApiReply{}
			restruct.Unpack(frame.data, binary.LittleEndian, &vv)
			return vv, nil
		case flashWriteApi:
			vv := FlashWriteApiReply{}
			restruct.Unpack(frame.data, binary.LittleEndian, &vv)
			return vv, nil
		case flashEBootApiRequest:
			vv := FlashEBootApiReply{}
			restruct.Unpack(frame.data, binary.LittleEndian, &vv)
			return vv, nil
		}
	}

	return nil, fmt.Errorf("unknow api frame: %d", frame.data[0])
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
	case NodeGetTagApiRequest:
		v.Id = nodeGetTagApiRequest
		b, err = restruct.Pack(binary.LittleEndian, &v)
	case NodeSetTagApiRequest:
		v.Id = nodeSetTagApiRequest
		b, err = restruct.Pack(binary.LittleEndian, &v)
	case NodeBindClearApiRequest:
		v.Id = nodeBindClearApiRequest
		b, err = restruct.Pack(binary.LittleEndian, &v)
	case NodeSetChannelApiRequest:
		v.Id = nodeSetChannelApiRequest
		b, err = restruct.Pack(binary.LittleEndian, &v)
	case NodeConfigApiRequest:
		v.Id = nodeConfigApiRequest
		b, err = restruct.Pack(binary.LittleEndian, &v)
	case NodeRebootApiRequest:
		v.Id = nodeRebootApiRequest
		b, err = restruct.Pack(binary.LittleEndian, &v)
	case EntitiesCountApiRequest:
		v.Id = entitiesCountApiRequest
		b, err = restruct.Pack(binary.LittleEndian, &v)
	case EntityHashApiRequest:
		v.Id = entityHashApiRequest
		b, err = restruct.Pack(binary.LittleEndian, &v)
	case GetEntityStateApiRequest:
		v.Id = getEntityStateApiRequest
		b, err = restruct.Pack(binary.LittleEndian, &v)
	case SetEntityStateApiRequest:
		v.Id = setEntityStateApiRequest
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
	case FlashGetMd5ApiRequest:
		v.Id = flashOperationApiRequest
		v.ApiId = flashGetMd5Api
		b, err = restruct.Pack(binary.LittleEndian, &v)
	case FlashEraseApiRequest:
		v.Id = flashOperationApiRequest
		v.ApiId = flashEraseApi
		b, err = restruct.Pack(binary.LittleEndian, &v)
	case FlashWriteApiRequest:
		v.Id = flashOperationApiRequest
		v.ApiId = flashWriteApi
		b, err = restruct.Pack(binary.LittleEndian, &v)
	case FlashEBootApiRequest:
		v.Id = flashOperationApiRequest
		v.ApiId = flashEBootApiRequest
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
			//frame.escaped = true
		}
	}

	return err
}

func FindBestProtocol(target MeshNodeId, network *graph.Network) MeshProtocol {
	if network == nil {
		return UnicastProtocol
	}

	device, err := network.GetNodeDevice(int64(target))
	if err != nil {
		return UnicastProtocol
	}

	if network.LocalDeviceId() == device.ID() {
		return DirectProtocol
	}

	path, _, err := network.GetPath(device)
	if err != nil {
		return UnicastProtocol
	}

	if len(path) == 1 {
		return DirectProtocol
	} else if len(path) == 2 {
		return UnicastProtocol
	} else {
		return MultipathProtocol
	}
}

func FindBestProtocolOverride(target MeshNodeId, protocol MeshProtocol, network *graph.Network) MeshProtocol {
	if protocol == AutoProtocol {
		return FindBestProtocol(target, network)
	}
	return protocol
}

func NewApiFrame(buffer []byte, escaped bool) *ApiFrame {
	f := &ApiFrame{
		data:    buffer,
		escaped: escaped,
	}

	return f
}

func NewApiFrameFromStruct(v interface{}, protocol MeshProtocol, target MeshNodeId, network *graph.Network) (*ApiFrame, error) {
	f := &ApiFrame{}

	switch protocol {
	case DirectProtocol:
		// direct prtocol talk with the serial connected device
		err := f.EncodeFrame(v)
		if err != nil {
			return nil, err
		}
	case UnicastProtocol:
		// unicast protocol talk with the mesh network without hops
		var err error
		p := UnicastRequest{Id: connectedUnicastRequest, Target: target}
		p.Payload, err = EncodeBuffer(v)
		if err != nil {
			return nil, err
		}
		err = f.EncodeFrame(p)
		if err != nil {
			return nil, err
		}
	case MultipathProtocol:
		// multipath protocol talk with the mesh network with hops
		if network == nil {
			return nil, errors.New("multipathProtocol requested, but network graph not initialized")
		}
		device, err := network.GetNodeDevice(int64(target))
		if err != nil {
			return nil, err
		}
		path, _, err := network.GetPath(device)
		if err != nil {
			return nil, err
		}
		if len(path) == 1 {
			return nil, errors.New("requested target is the local node. Use directProtocol instead")
		}
		// Removed the local node and the target node from the path
		_path := make([]uint32, len(path)-2)
		for i, p := range path[1 : len(path)-1] {
			_path[i] = uint32(p)
		}
		// Initialize the multipath request with the path and the target
		p := MultiPathRequest{Id: multipathRequest, Target: target, PathLen: uint8(len(_path)), Path: _path}
		p.Payload, err = EncodeBuffer(v)
		if err != nil {
			return nil, err
		}
		// Encode the multipath request
		err = f.EncodeFrame(p)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unknow protocol requested")
	}

	return f, nil
}
