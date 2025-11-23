package meshmesh

import (
	"leguru.net/m/v2/graph"
	"leguru.net/m/v2/logger"
	"leguru.net/m/v2/utils"
)

type StarPath struct {
	serial  *SerialConnection
	network *graph.Network
}

func (s *StarPath) GetNetwork() *graph.Network {
	return s.network
}

func (s *StarPath) handleNodePresentstionReply(v *NodePresentstionApiReply, serial *SerialConnection) {
	logger.WithFields(logger.Fields{"source": utils.FmtNodeId(int64(v.SourceAddr)), "target": utils.FmtNodeId(int64(v.TargetAddr))}).Info("NodePresentstionReply received")
	for i := range v.Hops {
		if v.Repeaters[i] > 0 {
			logger.WithFields(logger.Fields{"repeater": utils.FmtNodeId(int64(v.Repeaters[i])), "rssi": v.Rssi[i]}).Debug("NodePresentstionReply received")
		}
	}

	if uint32(v.TargetAddr) == serial.LocalNode {
		sourceNode := s.network.Node(int64(v.SourceAddr))
		if sourceNode == nil {
			sourceNode = graph.NewNodeDevice(int64(v.SourceAddr), true, "")
			s.network.AddNode(sourceNode)
		}

		for i := range v.Hops {
			node := s.network.Node(int64(v.Repeaters[i]))
			if node == nil {
				node = graph.NewNodeDevice(int64(v.Repeaters[i]), true, "")
				s.network.AddNode(node)
			}
			weight := Rssi2weight(v.Rssi[i])
			s.network.ChangeEdgeWeight(node.ID(), sourceNode.ID(), weight, weight)
			sourceNode = node
		}

		s.network.SaveToFile("starpath.graphml")
	}
}

func NewStarPath(serial *SerialConnection) *StarPath {
	starPath := &StarPath{
		serial:  serial,
		network: graph.NewNetwork(int64(serial.LocalNode)),
	}
	starPath.serial.NodePresentstionFn = starPath.handleNodePresentstionReply
	return starPath
}
