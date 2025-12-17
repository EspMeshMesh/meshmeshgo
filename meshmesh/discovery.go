package meshmesh

import (
	"errors"
	"math"
	"time"

	"leguru.net/m/v2/logger"
	"leguru.net/m/v2/utils"

	gra "leguru.net/m/v2/graph"
)

const maxRepetitions = 3

type DiscoveryProcedureState int

const (
	DiscoveryProcedureStateIdle DiscoveryProcedureState = iota
	DiscoveryProcedureStateRun
	DiscoveryProcedureStateDiscovering
	DiscoveryProcedureStateDone
	DiscoveryProcedureStateError
)

type DiscoveryProcedure struct {
	serial          *SerialConnection
	network         *gra.Network
	currentDeviceId int64
	Neighbors       map[int64]discWeights
	state           DiscoveryProcedureState
	repeat          int
}

func (d *DiscoveryProcedure) State() DiscoveryProcedureState {
	return d.state
}

func (d *DiscoveryProcedure) StateString() string {
	switch d.state {
	case DiscoveryProcedureStateIdle:
		return "idle"
	case DiscoveryProcedureStateRun:
		return "running"
	case DiscoveryProcedureStateDiscovering:
		return "discovering"
	case DiscoveryProcedureStateDone:
		return "done"
	case DiscoveryProcedureStateError:
		return "error"
	}
	return "unknown"
}

func (d *DiscoveryProcedure) CurrentDeviceId() int64 {
	return d.currentDeviceId
}

func (d *DiscoveryProcedure) CurrentRepeat() int {
	return d.repeat
}

type discWeights struct {
	Current float64
	Next    float64
}

const esp32RssiMax = -40
const esp32RssiMin = -80

func Rssi2weight(rssi int16) float64 {
	percent := 0.0
	if rssi <= 0 {
		percent = float64(esp32RssiMin-rssi) / float64(esp32RssiMin-esp32RssiMax)
	} else if rssi > 44 {
		percent = float64(rssi) / 45.0
		rssi = 44
	}

	cost := 1.0 - percent
	if cost < 0.0 {
		cost = 0.0
	} else if cost > 1.0 {
		cost = 1.0
	}

	return math.Round(cost*100) / 100
}

func CostToWeight(cost int16) float64 {
	if cost < 0.0 {
		return Rssi2weight(cost)
	} else if cost > 44 {
		return float64(cost)
	} else {
		return float64(cost)
	}
}

func neighborsFromGraph(g *gra.Network, n gra.NodeDevice, w map[int64]discWeights) error {
	neighbors := g.From(n.ID())
	for neighbors.Next() {
		neighbor := neighbors.Node().(gra.NodeDevice)
		weightTo, ok := g.Weight(n.ID(), neighbor.ID())
		if !ok {
			return errors.New("corrupted graph")
		}
		weightFrom, ok := g.Weight(neighbor.ID(), n.ID())
		if !ok {
			weightFrom = weightTo
			logger.WithFields(logger.Fields{"to": gra.FmtDeviceId(neighbor), "weightTo": weightTo, "weightFrom": weightFrom}).
				Warnf("[%s] Missing return edge", gra.FmtDeviceId(n))
		}
		w[neighbor.ID()] = discWeights{Next: math.Min(weightTo, weightFrom), Current: 1.0}
	}
	return nil
}

func _neighborsAdavance(w map[int64]discWeights) {
	for i, d := range w {
		w[i] = discWeights{Current: d.Next, Next: 1.0}
	}
}

func neighborsToGraph(g *gra.Network, nodeId int64, w map[int64]discWeights) {
	nodes := g.From(nodeId)

	for nodes.Next() {
		neighbor := nodes.Node().(gra.NodeDevice)
		g.RemoveEdge(nodeId, neighbor.ID())
	}

	for id, d := range w {
		logger.WithFields(logger.Fields{"to": utils.FmtNodeId(id), "weight": d, "exists": g.NodeIdExists(id)}).
			Infof("[%s] Neighbor to graph", utils.FmtNodeId(nodeId))
		g.ChangeEdgeWeight(nodeId, id, d.Next, d.Next)
	}
}

func _updateNeighbor(w map[int64]discWeights, id int64, rssi1, rssi2 float64) error {
	if _, exists := w[id]; exists {
		w[id] = discWeights{Current: w[id].Current, Next: math.Min(rssi1, rssi2)}
	} else {
		w[id] = discWeights{Current: 1.0, Next: math.Min(rssi1, rssi2)}
	}
	return nil
}

func _findNextNode(g *gra.Network) gra.NodeDevice {
	nodes := g.Nodes()
	var found_node gra.NodeDevice
	var found_weight float64 = 1e9
	for nodes.Next() {
		dev := nodes.Node().(gra.NodeDevice)
		if dev.Device().InUse() && !dev.Device().Discovered() {
			path, weight, err := g.GetPath(dev)

			if weight < found_weight {
				found_weight = weight
				found_node = dev
			}
			logger.WithFields(logger.Fields{"path": utils.FmtPath2Str(path), "weight": weight, "err": err}).
				Debugf("[%s] Not discovered node", utils.FmtNodeId(dev.ID()))
		}
	}
	return found_node
}

func (d *DiscoveryProcedure) InitStep() error {
	d.state = DiscoveryProcedureStateRun

	if d.network == nil {
		d.network = gra.NewNetwork(int64(d.serial.LocalNode), gra.NETWORK_ID_DISCOVERY)
	}

	if d.currentDeviceId != 0 {
		if d.repeat < maxRepetitions {
			// Repeat same node
			return nil
		} else {
			d.currentDeviceId = 0
		}
	}

	if d.currentDeviceId == 0 {
		d.currentDeviceId = _findNextNode(d.network).ID()
	}

	if d.currentDeviceId == 0 {
		d.state = DiscoveryProcedureStateDone
		return errors.New("no nodes to discover")
	}

	d.Neighbors = make(map[int64]discWeights)

	node, err := d.network.GetNodeDevice(d.currentDeviceId)
	if err != nil {
		return err
	}

	neighborsFromGraph(d.network, node, d.Neighbors)
	d.repeat = 0
	return nil
}

func (d *DiscoveryProcedure) Step() error {
	protocol := FindBestProtocol(MeshNodeId(d.currentDeviceId), d.network)
	logger.Log().Printf("[%s] Start discover with protocol %d repetition %d", utils.FmtNodeId(d.currentDeviceId), protocol, d.repeat)

	_, err := d.serial.SendReceiveApiProt(DiscResetTableApiRequest{}, protocol, MeshNodeId(d.currentDeviceId), d.network)
	if err != nil {
		return err
	}

	_, err = d.serial.SendReceiveApiProt(DiscStartDiscoverApiRequest{Mask: 0, Filter: 0, Slotnum: 100}, protocol, MeshNodeId(d.currentDeviceId), d.network)
	if err != nil {
		return err
	}

	// Get tag string from device and if the graph description is empty set the same as the tag on the device.
	_reply, err := d.serial.SendReceiveApiProt(NodeGetTagApiRequest{}, protocol, MeshNodeId(d.currentDeviceId), d.network)
	if err != nil {
		return err
	}
	tagReply, ok := _reply.(NodeGetTagApiReply)
	if !ok {
		return errors.New("strcuture is not a NodeGetTagApiReply")
	} else {
		logger.Log().Printf("[%s] Tag: %s", utils.FmtNodeId(d.currentDeviceId), tagReply.Tag)
	}

	_device, err := d.network.GetNodeDevice(d.currentDeviceId)
	if err != nil {
		return err
	}

	if len(_device.Device().Tag()) == 0 {
		_device.Device().SetTag(utils.TruncateZeros(tagReply.Tag))
	}

	time.Sleep(5 * time.Second)

	reply1, err := d.serial.SendReceiveApiProt(DiscTableSizeApiRequest{}, protocol, MeshNodeId(d.currentDeviceId), d.network)
	if err != nil {
		return err
	}
	tableSize, ok := reply1.(DiscTableSizeApiReply)
	if !ok {
		return errors.New("comunication error")
	}

	_neighborsAdavance(d.Neighbors)
	logger.Log().Printf("[%s] Discovered nodes: %d", utils.FmtNodeId(d.currentDeviceId), tableSize.Size)
	for i := uint8(0); i < tableSize.Size; i++ {

		reply1, err = d.serial.SendReceiveApiProt(DiscTableItemGetApiRequest{Index: i}, protocol, MeshNodeId(d.currentDeviceId), d.network)
		if err != nil {
			return err
		}
		tableItem, ok := reply1.(DiscTableItemGetApiReply)
		if !ok {
			return errors.New("comunication error")
		}

		logger.Log().Printf("         %d: [%s] rssi1 %d rssi2 %d", i, utils.FmtNodeId(int64(tableItem.NodeId)), tableItem.Rssi1, tableItem.Rssi2)
		_updateNeighbor(d.Neighbors, int64(tableItem.NodeId), Rssi2weight(tableItem.Rssi1), Rssi2weight(tableItem.Rssi2))
	}
	return err
}

func (d *DiscoveryProcedure) Clear() {
	if d.network != nil {
		nodes := d.network.Nodes()
		for nodes.Next() {
			node := nodes.Node().(gra.NodeDevice)
			node.Device().SetDiscovered(false)
		}
	}
	d.state = DiscoveryProcedureStateIdle
}

func (d *DiscoveryProcedure) Save() error {
	if d.currentDeviceId == 0 {
		return errors.New("discovery is inactive")
	}

	node, err := d.network.GetNodeDevice(d.currentDeviceId)
	if err != nil {
		return err
	}

	d.repeat++
	node.Device().SetDiscovered(true)
	neighborsToGraph(d.network, d.currentDeviceId, d.Neighbors)
	return nil
}

func (d *DiscoveryProcedure) Run() {
	d.Clear()
	for d.state != DiscoveryProcedureStateDone && d.state != DiscoveryProcedureStateError {
		d.InitStep()
		if d.state == DiscoveryProcedureStateRun {
			err := d.Step()
			if err != nil {
				d.state = DiscoveryProcedureStateError
				logger.Log().Println("Discovery procedure error", err)
			} else {
				d.Save()
			}
		}
	}

	gra.SetMainNetwork(d.network)
}

func NewDiscoveryProcedure(serial *SerialConnection, network *gra.Network, nodeid int64) *DiscoveryProcedure {
	return &DiscoveryProcedure{serial: serial, network: network, currentDeviceId: 0, state: DiscoveryProcedureStateIdle, repeat: 0}
}
