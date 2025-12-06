package meshmesh

import (
	"os"
	"time"

	"leguru.net/m/v2/graph"
	"leguru.net/m/v2/logger"
	pb "leguru.net/m/v2/meshmesh/pb"
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

func (s *StarPath) handleProtoPresentationRxReply(v *pb.NodePresentationRx, serial *SerialConnection) {
	if v.NodePresentation == nil {
		logger.Log().Error("NodePresentation is nil")
		return
	}
	if v.PathRouting == nil {
		logger.Log().Error("PathRouting is nil")
		return
	}

	if len(v.PathRouting.Rssi) < 1 && len(v.PathRouting.Rssi) <= len(v.PathRouting.Repeaters) {
		logger.Log().Error("PathRouting has not enough rssi data")
		return
	}

	logger.WithFields(logger.Fields{"source": utils.FmtNodeId(int64(v.PathRouting.SourceAddress)), "target": utils.FmtNodeId(int64(v.PathRouting.TargetAddress))}).Info("NodePresentstionReply received")
	logger.WithFields(logger.Fields{"hostname": v.NodePresentation.Hostname, "firmware": v.NodePresentation.FirmwareVersion, "compile_time": v.NodePresentation.CompileTime}).Info("ProtoPresentationReply received")
	for i := range len(v.PathRouting.Repeaters) {
		logger.WithFields(logger.Fields{"repeater": utils.FmtNodeId(int64(v.PathRouting.Repeaters[i])), "rssi": v.PathRouting.Rssi[i]}).Debug("NodePresentstionReply received")
	}

	if uint32(v.PathRouting.TargetAddress) == serial.LocalNode {
		sourceNode, err := s.network.GetNodeDevice(int64(v.PathRouting.SourceAddress))
		if err == nil {
			s.network.RemoveNode(sourceNode.ID())
		}
		sourceNode = graph.NewNodeDevice(int64(v.PathRouting.SourceAddress), true, "")
		sourceNode.Device().SetTag(v.NodePresentation.Hostname)
		sourceNode.Device().SetFirmware(v.NodePresentation.FirmwareVersion)
		sourceNode.Device().SetCompileTimeString(v.NodePresentation.CompileTime)
		sourceNode.Device().SetLastSeen(time.Now())
		s.network.AddNode(sourceNode)

		v.PathRouting.Repeaters = append(v.PathRouting.Repeaters, v.PathRouting.TargetAddress)

		for i := range len(v.PathRouting.Repeaters) {
			node, err := s.network.GetNodeDevice(int64(v.PathRouting.Repeaters[i]))
			if err != nil {
				node = graph.NewNodeDevice(int64(v.PathRouting.Repeaters[i]), true, "")
				s.network.AddNode(node)
			}
			if !s.network.HasEdgeFromTo(node.ID(), sourceNode.ID()) {
				weight := Rssi2weight(int16(v.PathRouting.Rssi[i]))
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
func initNetwork(localNodeId int64) *graph.Network {
	var network *graph.Network
	if _, err := os.Stat(starPathGraphFilename); err == nil {
		network, err = graph.NewNeworkFromFile(starPathGraphFilename, localNodeId, graph.NETWORK_ID_STARPATH)
		if err != nil {
			logger.Log().Fatal("Graph read error: ", err)
		}
	} else {
		network = graph.NewNetwork(localNodeId, graph.NETWORK_ID_STARPATH)
		network.SaveToFile(starPathGraphFilename)
	}
	return network
}

func NewStarPath(serial *SerialConnection) *StarPath {
	starPath := &StarPath{
		serial:  serial,
		network: initNetwork(int64(serial.LocalNode)),
	}
	starPath.serial.ProtoPresentationFn = starPath.handleProtoPresentationRxReply
	return starPath
}
