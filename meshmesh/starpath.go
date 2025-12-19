package meshmesh

import (
	"os"
	"slices"
	"time"

	"leguru.net/m/v2/graph"
	"leguru.net/m/v2/logger"
	pb "leguru.net/m/v2/meshmesh/pb"
	"leguru.net/m/v2/utils"
)

type StarPath struct {
	serial  *SerialConnection
	network *graph.Network
}

func (s *StarPath) GetNetwork() *graph.Network {
	return s.network
}

/*
If toId not exists, create a new node with the toId and add it to the network, otherwise remove all input edges from the toId node.
Then add a new edge from the fromId node to the toId node.
*/
func (s *StarPath) refreshInputEdges(fromId int64, toId int64, weight float64) {
	if !s.network.NodeIdExists(toId) {
		node := graph.NewNodeDevice(toId, true, "")
		s.network.AddNode(node)
	} else {
		edges := s.network.EdgesTo(toId)
		for edges.Next() {
			edge := edges.Edge()
			s.network.RemoveEdge(edge.From().ID(), edge.To().ID())
		}
	}
	s.network.ChangeEdgeWeight(fromId, toId, weight, weight)
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

	if v.PathRouting.TargetAddress != uint32(serial.LocalNode) {
		logger.Log().Error("PathRouting target address is not the local node")
		return
	}

	logger.WithFields(logger.Fields{"source": utils.FmtNodeId(int64(v.PathRouting.SourceAddress)), "target": utils.FmtNodeId(int64(v.PathRouting.TargetAddress))}).Info("NodePresentstionReply received")
	logger.WithFields(logger.Fields{"hostname": v.NodePresentation.Hostname, "firmware": v.NodePresentation.FirmwareVersion, "compile_time": v.NodePresentation.CompileTime, "lib_version": v.NodePresentation.LibVersion}).Info("ProtoPresentationReply received")
	for i := range len(v.PathRouting.Repeaters) {
		logger.WithFields(logger.Fields{"repeater": utils.FmtNodeId(int64(v.PathRouting.Repeaters[i])), "rssi": v.PathRouting.Rssi[i]}).Debug("NodePresentstionReply received")
	}

	if uint32(v.PathRouting.TargetAddress) == serial.LocalNode {
		sourceNodeIsNew := false

		sourceNode, err := s.network.GetNodeDevice(int64(v.PathRouting.SourceAddress))
		if err != nil {
			sourceNode = graph.NewNodeDevice(int64(v.PathRouting.SourceAddress), true, "")
			sourceNodeIsNew = true
		}
		sourceNode.Device().SetTag(v.NodePresentation.Hostname)
		sourceNode.Device().SetFirmware(v.NodePresentation.FirmwareVersion)
		sourceNode.Device().SetCompileTimeString(v.NodePresentation.CompileTime)
		sourceNode.Device().SetLibVersion(v.NodePresentation.LibVersion)
		sourceNode.Device().SetLastSeen(time.Now())

		if sourceNodeIsNew {
			s.network.AddNode(sourceNode)
		}

		// Create the path from Target (coordinator) to Source (node)
		slices.Reverse(v.PathRouting.Repeaters)
		path := append([]uint32{v.PathRouting.TargetAddress}, v.PathRouting.Repeaters...)
		path = append(path, v.PathRouting.SourceAddress)

		for i := range len(path) - 1 {
			// new edge is: from:node[i] -> rssi[i] --> to:node[i+1]
			s.refreshInputEdges(int64(path[i]), int64(path[i+1]), CostToWeight(int16(v.PathRouting.Rssi[i])))
		}

		s.network.NotifyNetworkChanged()
	}
}

/*
Init star path network graph from cache file or create a new one if not exists
*/
func initNetwork(localNodeId int64, filename string) *graph.Network {
	var network *graph.Network
	if _, err := os.Stat(filename); err == nil {
		network, err = graph.NewNeworkFromFile(filename, localNodeId, graph.NETWORK_ID_STARPATH)
		if err != nil {
			logger.Log().Fatal("Graph read error: ", err)
		}
	} else {
		network = graph.NewNetwork(localNodeId, graph.NETWORK_ID_STARPATH)
		network.SaveToFile(filename)
	}
	return network
}

func NewStarPath(serial *SerialConnection, cacheFile string) *StarPath {
	starPath := &StarPath{
		serial:  serial,
		network: initNetwork(int64(serial.LocalNode), cacheFile),
	}
	starPath.serial.ProtoPresentationFn = starPath.handleProtoPresentationRxReply
	return starPath
}
