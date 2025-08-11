package graph

import (
	"fmt"
	"math"
	"sync"

	"github.com/sirupsen/logrus"
	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
	"leguru.net/m/v2/logger"
	"leguru.net/m/v2/utils"
)

type Device struct {
	inuse      bool
	discovered bool
	tag        string
}

func (d Device) InUse() bool {
	return d.inuse
}

func (d *Device) SetInUse(inuse bool) {
	d.inuse = inuse
}

func (d Device) Discovered() bool {
	return d.discovered
}

func (d *Device) SetDiscovered(discovered bool) {
	d.discovered = discovered
}

func (d Device) Tag() string {
	return d.tag
}

func (d *Device) SetTag(tag string) {
	d.tag = tag
}

func NewDevice(inuse bool, tag string) *Device {
	return &Device{inuse: inuse, tag: tag}
}

type NodeDevice struct {
	id     int64
	device *Device
}

func (n NodeDevice) ID() int64 {
	return n.id
}

func (n NodeDevice) Device() *Device {
	return n.device
}

func (n NodeDevice) CopyDevice() NodeDevice {
	return NodeDevice{id: n.id, device: n.device}
}

func NewNodeDevice(id int64, inuse bool, tag string) NodeDevice {
	return NodeDevice{id: id, device: NewDevice(inuse, tag)}
}

// Network: is a weighted directed graph of NodeDevices
var mainNetwork *Network
var mainNetworkChancgedCallbacks []func()
var mainNetworkLock sync.Mutex

func GetMainNetwork() *Network {
	mainNetworkLock.Lock()
	defer mainNetworkLock.Unlock()
	return mainNetwork
}

// SetMainNetwork sets the current main network instance pointer and notify all callbacks.
// It acquires a lock to ensure thread-safe access to the global mainNetwork variable.
func SetMainNetwork(network *Network) {
	mainNetworkLock.Lock()
	mainNetwork = network
	mainNetworkLock.Unlock()
	NotifyMainNetworkChanged()
}

func AddMainNetworkChangedCallback(cb func()) {
	mainNetworkChancgedCallbacks = append(mainNetworkChancgedCallbacks, cb)
}

func NotifyMainNetworkChanged() {
	for _, cb := range mainNetworkChancgedCallbacks {
		cb()
	}
}

type Network struct {
	simple.WeightedDirectedGraph
	localDeviceId int64
}

func (g *Network) LocalDeviceId() int64 {
	return g.localDeviceId
}

func (g *Network) GetNodeDevice(id int64) (NodeDevice, error) {
	if node, ok := g.Node(id).(NodeDevice); ok {
		return node, nil
	}
	return NodeDevice{}, fmt.Errorf("node 0x%06X not found in network graph", id)
}

func (g *Network) AddNodeWithId(id int64, inuse bool, tag string, seen bool) {
	g.AddNode(NewNodeDevice(id, inuse, tag))
}

func (g *Network) NodeIdExists(id int64) bool {
	return g.Node(id) != nil
}

func (g *Network) ChangeEdgeWeight(fromId int64, toId int64, weightFrom float64, weightTo float64) {
	fromNode, err := g.GetNodeDevice(fromId)
	if err != nil {
		fromNode = NewNodeDevice(fromId, false, "")
		g.AddNode(fromNode)
	}

	toNode, err := g.GetNodeDevice(toId)
	if err != nil {
		toNode = NewNodeDevice(toId, true, "")
		g.AddNode(toNode)
	}

	if !g.HasEdgeFromTo(fromId, toId) {
		g.SetWeightedEdge(g.NewWeightedEdge(fromNode, toNode, weightTo))
	} else {
		edgeTo := g.WeightedEdge(fromId, toId).(simple.WeightedEdge)
		edgeTo.W = weightTo
		g.SetWeightedEdge(edgeTo)
	}
}

// GetPath returns the shortest path from the local device to the target device, along with the total path weight.
//
// Parameters:
//   - to: The target Device to find a path to
//
// Returns:
//   - []int64: Array of node IDs representing the path from local device to target
//   - float64: Total weight/cost of the path
//   - error: Error if no path exists or target device is not active
//
// The path returned will be the shortest path based on edge weights using Dijkstra's algorithm.
// Returns an error if:
// - The target device is not marked as in use/active
// - No valid path exists between the local device and target
// The weight returned is currently always 0 (not implemented).

func (g *Network) GetPath(to NodeDevice) ([]int64, float64, error) {
	if !to.Device().InUse() {
		return nil, 0, fmt.Errorf("node is 0x%06X is not active", to.ID())
	}
	allShortest := path.DijkstraAllPaths(g)
	allBetween, weight := allShortest.AllBetween(g.localDeviceId, to.ID())
	if len(allBetween) == 0 {
		return nil, 0, fmt.Errorf("no path found between 0x%06X and 0x%06X", g.localDeviceId, to.ID())
	}
	logrus.WithFields(logrus.Fields{"length": len(allBetween[0]), "weight": weight}).
		Debug(fmt.Sprintf("Get path from 0x%06X to 0x%06X", g.localDeviceId, to.ID()))

	nodes := allBetween[0]
	path := make([]int64, len(nodes))
	for i, item := range nodes {
		item := item.(NodeDevice)
		path[i] = item.ID()
	}

	return path, 0, nil
}

func (g *Network) SaveToFile(filename string) error {
	utils.BackupFile(filename, "backup")
	return g.writeGraph(filename)
}

func (g *Network) CopyNetwork() *Network {
	network := Network{}
	network.WeightedDirectedGraph = *simple.NewWeightedDirectedGraph(0, math.Inf(1))
	network.localDeviceId = g.localDeviceId

	nodes := g.Nodes()
	for nodes.Next() {
		dev := nodes.Node().(NodeDevice)
		network.AddNode(dev.CopyDevice())
	}

	edges := g.Edges()
	for edges.Next() {
		edge := edges.Edge().(simple.WeightedEdge)
		network.SetWeightedEdge(g.NewWeightedEdge(edge.From(), edge.To(), edge.Weight()))
	}

	return &network
}

func NewNetwork(localDeviceId int64) *Network {
	network := Network{localDeviceId: localDeviceId}
	network.WeightedDirectedGraph = *simple.NewWeightedDirectedGraph(0, math.Inf(1))
	network.AddNode(NewNodeDevice(localDeviceId, true, "local"))
	return &network
}

func NewNeworkFromFile(filename string, localDeviceId int64) (*Network, error) {
	network := Network{}
	network.WeightedDirectedGraph = *simple.NewWeightedDirectedGraph(0, math.Inf(1))
	err := network.readGraph(filename)
	if err != nil {
		return nil, err
	}

	network.localDeviceId = localDeviceId
	if !network.NodeIdExists(localDeviceId) {
		network.AddNode(NewNodeDevice(localDeviceId, true, "local"))
		logger.WithField("device", utils.FmtNodeId(localDeviceId)).Warn("Local device not found in graph, adding it. Will be an isolated node")
	}
	return &network, nil
}
