package meshmesh

import (
	"os"

	"leguru.net/m/v2/graph"
	gra "leguru.net/m/v2/graph"
	"leguru.net/m/v2/logger"
	"leguru.net/m/v2/utils"
)

const (
	starPathGraphFilename = "starpath.graphml"
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
	logger.WithFields(logger.Fields{"firmware": utils.TruncateZeros(v.FwVersion[:]), "hostname": utils.TruncateZeros(v.Hostname[:]), "compile_time": utils.TruncateZeros(v.CompileTime[:])}).Debug("NodePresentstionReply received")
	for i := range v.Hops {
		if v.Repeaters[i] > 0 {
			logger.WithFields(logger.Fields{"repeater": utils.FmtNodeId(int64(v.Repeaters[i])), "rssi": v.Rssi[i]}).Debug("NodePresentstionReply received")
		}
	}

	if uint32(v.TargetAddr) == serial.LocalNode {
		sourceNode, err := s.network.GetNodeDevice(int64(v.SourceAddr))
		if err == nil {
			s.network.RemoveNode(sourceNode.ID())
		}
		sourceNode = graph.NewNodeDevice(int64(v.SourceAddr), true, "")
		sourceNode.Device().SetTag(utils.TruncateZeros(v.Hostname[:]))
		sourceNode.Device().SetFirmware(utils.TruncateZeros(v.FwVersion[:]))
		sourceNode.Device().SetCompileTime(utils.TruncateZeros(v.CompileTime[:]))
		s.network.AddNode(sourceNode)

		for i := range v.Hops {
			node, err := s.network.GetNodeDevice(int64(v.Repeaters[i]))
			if err != nil {
				node = graph.NewNodeDevice(int64(v.Repeaters[i]), true, "")
				s.network.AddNode(node)
			}
			if !s.network.HasEdgeFromTo(node.ID(), sourceNode.ID()) {
				weight := Rssi2weight(v.Rssi[i])
				s.network.ChangeEdgeWeight(node.ID(), sourceNode.ID(), weight, weight)
			}
			sourceNode = node
		}

		s.network.SaveToFile("starpath.graphml")
	}
}

/*
Init star path network graph from cache file or create a new one if not exists
*/
func initNetwork(localNodeId int64) *gra.Network {
	var network *gra.Network
	if _, err := os.Stat(starPathGraphFilename); err == nil {
		network, err = graph.NewNeworkFromFile(starPathGraphFilename, localNodeId, gra.NETWORK_ID_STARPATH)
		if err != nil {
			logger.Log().Fatal("Graph read error: ", err)
		}
	} else {
		network = gra.NewNetwork(localNodeId, gra.NETWORK_ID_STARPATH)
		network.SaveToFile(starPathGraphFilename)
	}
	return network
}

func NewStarPath(serial *SerialConnection) *StarPath {
	starPath := &StarPath{
		serial:  serial,
		network: initNetwork(int64(serial.LocalNode)),
	}
	starPath.serial.NodePresentstionFn = starPath.handleNodePresentstionReply
	return starPath
}
