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

func Rssi2weight(rssi int16) float64 {
	if rssi <= 0 {
		rssi *= -2
	} else if rssi > 44 {
		rssi = 44
	}
	cost := 1.0 - float64(rssi)/45.0
	return math.Round(cost*100) / 100
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
			logger.WithFields(logger.Fields{"from": gra.FmtDeviceId(n), "to": gra.FmtDeviceId(neighbor), "weightTo": weightTo, "weightFrom": weightFrom}).Warn("Missing return edge")
			//return errors.New("corrupted graph")
		}
		w[neighbor.ID()] = discWeights{Current: math.Min(weightTo, weightFrom), Next: 1.0}
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
		logger.WithFields(logger.Fields{"from": nodeId, "to": id, "weight": d, "exists": g.NodeIdExists(id)}).Info("neighbor to graph")
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
			logger.WithFields(logger.Fields{"node": dev.ID(), "path": path, "weight": weight, "err": err}).Debug("_findNextNode not disvoered node")
		}
	}
	return found_node
}

func (d *DiscoveryProcedure) Init(forever bool) error {
	d.state = DiscoveryProcedureStateRun

	if d.network == nil {
		d.network = gra.NewNetwork(int64(d.serial.LocalNode))
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
	if d.currentDeviceId == 0 && forever {
		d.Clear()
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
	logger.Log().Printf("Start dicover of node 0x%06X with protocol %d repetition %d", d.currentDeviceId, protocol, d.repeat)

	_, err := d.serial.SendReceiveApiProt(DiscResetTableApiRequest{}, protocol, MeshNodeId(d.currentDeviceId), d.network)
	if err != nil {
		return err
	}

	_, err = d.serial.SendReceiveApiProt(DiscStartDiscoverApiRequest{Mask: 0, Filter: 0, Slotnum: 100}, protocol, MeshNodeId(d.currentDeviceId), d.network)
	if err != nil {
		return err
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
	logger.Log().Printf("Discovery of node 0x%06X: table size %d", d.currentDeviceId, tableSize.Size)
	for i := uint8(0); i < tableSize.Size; i++ {

		reply1, err = d.serial.SendReceiveApiProt(DiscTableItemGetApiRequest{Index: i}, protocol, MeshNodeId(d.currentDeviceId), d.network)
		if err != nil {
			return err
		}
		tableItem, ok := reply1.(DiscTableItemGetApiReply)
		if !ok {
			return errors.New("comunication error")
		}

		logger.Log().Printf("Query of row %d node %s rssi1 %d rssi2 %d", i, utils.FmtNodeId(int64(tableItem.NodeId)), tableItem.Rssi1, tableItem.Rssi2)
		_updateNeighbor(d.Neighbors, int64(tableItem.NodeId), Rssi2weight(tableItem.Rssi1), Rssi2weight(tableItem.Rssi2))
	}

	logger.Log().Printf("Discovery of node 0x%06X done", d.currentDeviceId)
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
	gra.SetMainNetwork(d.network)
	gra.NotifyMainNetworkChanged()
	return nil
}

func (d *DiscoveryProcedure) Run() {
	first := true
	for d.state != DiscoveryProcedureStateDone && d.state != DiscoveryProcedureStateError {
		d.Init(first)
		first = false

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
}

func NewDiscoveryProcedure(serial *SerialConnection, network *gra.Network, nodeid int64) *DiscoveryProcedure {
	return &DiscoveryProcedure{serial: serial, network: network, currentDeviceId: 0, state: DiscoveryProcedureStateIdle, repeat: 0}
}
