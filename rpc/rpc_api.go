package rpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"leguru.net/m/v2/graph"
	mm "leguru.net/m/v2/meshmesh"
	"leguru.net/m/v2/rpc/meshmesh"
	"leguru.net/m/v2/utils"
)

func (s *Server) NodeInfo(_ context.Context, req *meshmesh.NodeInfoRequest) (*meshmesh.NodeInfoReply, error) {
	mmid := mm.MeshNodeId(req.Id)
	network := graph.GetMainNetwork()
	rep, err := s.serialConn.SendReceiveApiProt(mm.FirmRevApiRequest{}, mm.FindBestProtocol(mmid, network), mmid, network)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get firmware revision: %v", err)
	}
	rev := rep.(mm.FirmRevApiReply)

	rep, err = s.serialConn.SendReceiveApiProt(mm.NodeConfigApiRequest{}, mm.UnicastProtocol, mm.MeshNodeId(req.Id), network)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get node configuration: %v", err)
	}
	cfg := rep.(mm.NodeConfigApiReply)

	return &meshmesh.NodeInfoReply{
		Id:           req.Id,
		Tag:          utils.TruncateZeros(cfg.Tag),
		Channel:      uint32(cfg.Channel),
		Rev:          utils.TruncateZeros(rev.Revision),
		IsAssociated: cfg.Flags&0x01 != 0,
	}, nil
}

func (s *Server) NodeReboot(_ context.Context, req *meshmesh.NodeRebootRequest) (*meshmesh.NodeRebootReply, error) {
	mmid := mm.MeshNodeId(req.Id)
	network := graph.GetMainNetwork()
	_, err := s.serialConn.SendReceiveApiProt(mm.NodeRebootApiRequest{}, mm.FindBestProtocol(mmid, network), mmid, network)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to reboot node: %v", err)
	}
	return &meshmesh.NodeRebootReply{Success: true}, nil
}

func (s *Server) BindClear(_ context.Context, req *meshmesh.BindClearRequest) (*meshmesh.BindClearReply, error) {
	mmid := mm.MeshNodeId(req.Id)
	network := graph.GetMainNetwork()
	_, err := s.serialConn.SendReceiveApiProt(mm.NodeBindClearApiRequest{}, mm.FindBestProtocol(mmid, network), mmid, network)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to clear binded server: %v", err)
	}
	return &meshmesh.BindClearReply{Success: true}, nil
}

func (s *Server) SetTag(_ context.Context, req *meshmesh.SetTagRequest) (*meshmesh.SetTagReply, error) {
	if len(req.Tag) > 30 {
		return nil, status.Errorf(codes.InvalidArgument, "Tag must be less than 30 characters")
	}
	mmid := mm.MeshNodeId(req.Id)
	network := graph.GetMainNetwork()
	_, err := s.serialConn.SendReceiveApiProt(mm.NodeSetTagApiRequest{Tag: req.Tag}, mm.FindBestProtocol(mmid, network), mmid, network)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to set tag: %v", err)
	}
	return &meshmesh.SetTagReply{Success: true}, nil
}

func (s *Server) SetChannel(_ context.Context, req *meshmesh.SetChannelRequest) (*meshmesh.SetChannelReply, error) {
	if req.Channel < 1 || req.Channel > 13 {
		return nil, status.Errorf(codes.InvalidArgument, "Channel must be between 1 and 13")
	}
	mmid := mm.MeshNodeId(req.Id)
	network := graph.GetMainNetwork()
	_, err := s.serialConn.SendReceiveApiProt(mm.NodeSetChannelApiRequest{Channel: uint8(req.Channel)}, mm.FindBestProtocol(mmid, network), mmid, network)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to set channel: %v", err)
	}
	return &meshmesh.SetChannelReply{Success: true}, nil
}

func (s *Server) EntitiesCount(_ context.Context, req *meshmesh.EntitiesCountRequest) (*meshmesh.EntitiesCountReply, error) {
	mmid := mm.MeshNodeId(req.Id)
	network := graph.GetMainNetwork()
	rep, err := s.serialConn.SendReceiveApiProt(mm.EntitiesCountApiRequest{}, mm.FindBestProtocol(mmid, network), mmid, network)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get entities count: %v", err)
	}
	cnt := rep.(mm.EntitiesCountApiReply)
	return &meshmesh.EntitiesCountReply{
		All:           uint32(cnt.Counters[0]),
		Sensors:       uint32(cnt.Counters[1]),
		BinarySensors: uint32(cnt.Counters[2]),
		Switches:      uint32(cnt.Counters[3]),
		Lights:        uint32(cnt.Counters[4]),
		TextSensors:   uint32(cnt.Counters[5]),
	}, nil
}

func (s *Server) EntityHash(_ context.Context, req *meshmesh.EntityHashRequest) (*meshmesh.EntityHashReply, error) {
	mmid := mm.MeshNodeId(req.Id)
	network := graph.GetMainNetwork()
	rep, err := s.serialConn.SendReceiveApiProt(mm.EntityHashApiRequest{Service: uint8(req.Service), Index: uint8(req.Index)}, mm.FindBestProtocol(mmid, network), mmid, network)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get entity hash: %v", err)
	}
	hash := rep.(mm.EntityHashApiReply)
	if hash.Hash == 0 && hash.Info == "E!" {
		return nil, status.Errorf(codes.NotFound, "Entity not found")
	}
	return &meshmesh.EntityHashReply{Id: req.Id, Hash: uint32(hash.Hash), Info: hash.Info}, nil
}

func (s *Server) GetEntityState(_ context.Context, req *meshmesh.GetEntityStateRequest) (*meshmesh.GetEntityStateReply, error) {
	mmid := mm.MeshNodeId(req.Id)
	network := graph.GetMainNetwork()
	rep, err := s.serialConn.SendReceiveApiProt(mm.GetEntityStateApiRequest{
		Service: uint8(req.Service),
		Hash:    uint16(req.Hash),
	}, mm.FindBestProtocol(mmid, network), mmid, network)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get entity state: %v", err)
	}
	state := rep.(mm.GetEntityStateApiReply)
	return &meshmesh.GetEntityStateReply{State: uint32(state.State)}, nil
}

func (s *Server) SetEntityState(_ context.Context, req *meshmesh.SetEntityStateRequest) (*meshmesh.SetEntityStateReply, error) {
	mmid := mm.MeshNodeId(req.Id)
	network := graph.GetMainNetwork()
	_, err := s.serialConn.SendReceiveApiProt(mm.SetEntityStateApiRequest{
		Service: uint8(req.Service),
		Hash:    uint16(req.Hash),
		State:   uint16(req.State),
	}, mm.FindBestProtocol(mmid, network), mmid, network)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to set entity state: %v", err)
	}
	return &meshmesh.SetEntityStateReply{Success: true}, nil
}

func (s *Server) ExecuteDiscovery(_ context.Context, req *meshmesh.ExecuteDiscoveryRequest) (*meshmesh.ExecuteDiscoveryReply, error) {
	mmid := mm.MeshNodeId(req.Id)
	network := graph.GetMainNetwork()
	_, err := s.serialConn.SendReceiveApiProt(mm.DiscStartDiscoverApiRequest{Mask: 0, Filter: 0, Slotnum: 100}, mm.FindBestProtocol(mmid, network), mmid, network)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to set entity state: %v", err)
	}
	return &meshmesh.ExecuteDiscoveryReply{Success: true}, nil
}
