package rpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"leguru.net/m/v2/graph"
	"leguru.net/m/v2/rpc/meshmesh"
)

func (s *Server) NetworkNodes(_ context.Context, req *meshmesh.NetworkNodesRequest) (*meshmesh.NetworkNodesReply, error) {
	nodes := graph.GetMainNetwork().Nodes()
	device := make([]*meshmesh.NetworkNode, nodes.Len())
	i := 0
	for nodes.Next() {
		dev := nodes.Node().(graph.NodeDevice)
		device[i] = &meshmesh.NetworkNode{
			Id:    uint32(dev.ID()),
			Tag:   string(dev.Device().Tag()),
			Inuse: dev.Device().InUse(),
		}
		i += 1
	}
	return &meshmesh.NetworkNodesReply{Nodes: device}, nil
}

func (s *Server) NetworkEdges(_ context.Context, req *meshmesh.NetworkEdgesRequest) (*meshmesh.NetworkEdgesReply, error) {
	edges := graph.GetMainNetwork().WeightedEdges()
	_edges := make([]*meshmesh.NetworkEdge, edges.Len())
	i := 0
	for edges.Next() {
		edge := edges.WeightedEdge()
		_edges[i] = &meshmesh.NetworkEdge{
			From:   uint32(edge.From().ID()),
			To:     uint32(edge.To().ID()),
			Weight: float32(edge.Weight()),
		}
		i += 1
	}
	return &meshmesh.NetworkEdgesReply{Edges: _edges}, nil
}

func (s *Server) NetworkNodeConfigure(_ context.Context, req *meshmesh.NetworkNodeConfigureRequest) (*meshmesh.NetworkNodeConfigureReply, error) {
	network := graph.GetMainNetwork()
	node := network.Node(int64(req.Id))
	if node == nil {
		return nil, status.Errorf(codes.NotFound, "Node not found")
	}
	dev := node.(graph.NodeDevice)
	dev.Device().SetTag(req.Tag)
	dev.Device().SetInUse(req.Inuse)
	network.NotifyNetworkChanged()
	return &meshmesh.NetworkNodeConfigureReply{Success: true}, nil
}

func (s *Server) NetworkNodeDelete(_ context.Context, req *meshmesh.NetworkNodeDeleteRequest) (*meshmesh.NetworkNodeDeleteReply, error) {
	network := graph.GetMainNetwork()
	node := network.Node(int64(req.Id))
	if node == nil {
		return nil, status.Errorf(codes.NotFound, "Node not found")
	}

	network.RemoveNode(int64(req.Id))
	network.NotifyNetworkChanged()
	return &meshmesh.NetworkNodeDeleteReply{Success: true}, nil
}
