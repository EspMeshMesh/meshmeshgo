package graph

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/iterator"
	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
	"leguru.net/m/v2/logger"
	"leguru.net/m/v2/meshmesh/pb"
	"leguru.net/m/v2/utils"
)

type NodeType int

const (
	NodeTypeBackbone NodeType = iota
	NodeTypeCoordinator
	NodeTypeEdge
)

func stringNodeTypeToEnum(nodeType string) NodeType {
	if nodeType == "backbone" {
		return NodeTypeBackbone
	}
	if nodeType == "coordinator" {
		return NodeTypeCoordinator
	}
	if nodeType == "edge" {
		return NodeTypeEdge
	}
	return NodeTypeBackbone
}

func EnumNodeTypeToString(nodeType NodeType) string {
	switch nodeType {
	case NodeTypeBackbone:
		return "backbone"
	case NodeTypeCoordinator:
		return "coordinator"
	case NodeTypeEdge:
		return "edge"
	}
	return "backbone"
}

const (
	NETWORK_ID_MAIN      = 0
	NETWORK_ID_STARPATH  = 1
	NETWORK_ID_DISCOVERY = 2
)

const (
	compileTimeFormat      = "Jan _2 2006, 15:04:05"
	compileTimeFormatShort = "Jan _2 2006"
	compileTimeFormatLong  = "2006-01-02 15:04:05 -0700"
)

type Device struct {
	inuse        bool
	deepSleep    bool
	discovered   bool
	tag          string
	nodeType     NodeType
	name         string
	friendlyName string
	firmware     string
	libVersion   string
	compileTime  time.Time
	lastSeen     time.Time
}

func (d *Device) InUse() bool {
	return d.inuse
}

func (d *Device) SetInUse(inuse bool) {
	d.inuse = inuse
}

func (d *Device) DeepSleep() bool {
	return d.deepSleep
}

func (d *Device) SetDeepSleep(deepSleep bool) {
	d.deepSleep = deepSleep
}

func (d *Device) Discovered() bool {
	return d.discovered
}

func (d *Device) SetDiscovered(discovered bool) {
	d.discovered = discovered
}

func (d *Device) Tag() string {
	return d.tag
}

func (d *Device) SetTag(tag string) {
	d.tag = tag
}

func (d *Device) NodeType() NodeType {
	return d.nodeType
}

func (d *Device) NodeTypeString() string {
	return EnumNodeTypeToString(d.nodeType)
}

func (d *Device) SetNodeType(nodeType NodeType) {
	d.nodeType = nodeType
}

func (d *Device) SetNodeTypeString(nodeType string) {
	d.nodeType = stringNodeTypeToEnum(nodeType)
}

func (d *Device) Name() string {
	return d.name
}

func (d *Device) SetName(name string) {
	d.name = name
}

func (d *Device) FriendlyName() string {
	return d.friendlyName
}

func (d *Device) SetFriendlyName(friendlyName string) {
	d.friendlyName = friendlyName
}

func (d *Device) Firmware() string {
	return d.firmware
}

func (d *Device) SetFirmware(firmware string) {
	d.firmware = firmware
}

func (d *Device) CompileTime() time.Time {
	return d.compileTime
}

func (d *Device) CompileTimeString() string {
	if d.compileTime.IsZero() {
		return ""
	}
	return d.compileTime.Format(compileTimeFormat)
}

func (d *Device) SetCompileTime(compileTime time.Time) {
	d.compileTime = compileTime
}

func (d *Device) SetCompileTimeString(compileTime string) {
	var err error

	d.compileTime, err = time.Parse(compileTimeFormatLong, compileTime)
	if err != nil {
		d.compileTime, err = time.Parse(compileTimeFormat, compileTime)
		if err != nil {
			sep := strings.Index(compileTime, ",")
			if sep != -1 {
				compileTime = compileTime[:sep]
			}
			d.compileTime, err = time.Parse(compileTimeFormatShort, compileTime)
			if err != nil {
				d.compileTime = time.Time{}
			}
		}
	}

}

func (d *Device) LibVersion() string {
	return d.libVersion
}

func (d *Device) SetLibVersion(libVersion string) {
	d.libVersion = libVersion
}

func (d *Device) LastSeen() time.Time {
	return d.lastSeen
}

func (d *Device) SetLastSeen(lastSeen time.Time) {
	d.lastSeen = lastSeen
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

func (n NodeDevice) DeviceTagOrFormattedId() string {
	if n.device != nil && n.device.tag != "" {
		return n.device.tag
	}
	return utils.FmtNodeId(n.id)
}

func (n NodeDevice) CopyDevice() NodeDevice {
	return NodeDevice{id: n.id, device: n.device}
}

func NewNodeDevice(id int64, inuse bool, tag string) NodeDevice {
	return NodeDevice{id: id, device: NewDevice(inuse, tag)}
}

// Network: is a weighted directed graph of NodeDevices
var mainNetwork *Network
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
	mainNetwork.networkId = NETWORK_ID_MAIN
	mainNetworkLock.Unlock()
	network.NotifyNetworkChanged()
}

type Network struct {
	simple.WeightedDirectedGraph
	localDeviceId           int64
	networkChangedCallbacks []func(network *Network)
	networkId               int
}

func (g *Network) EdgesTo(nodeId int64) graph.Edges {
	foundEdges := make([]graph.Edge, 0)

	edges := g.Edges()
	for edges.Next() {
		edge := edges.Edge()
		if edge.To().ID() == nodeId {
			foundEdges = append(foundEdges, edge)
		}
	}

	return iterator.NewOrderedEdges(foundEdges)
}

func (g *Network) NetworkId() int {
	return g.networkId
}

func (g *Network) AddNetworkChangedCallback(cb func(network *Network)) {
	g.networkChangedCallbacks = append(g.networkChangedCallbacks, cb)
}

func (g *Network) NotifyNetworkChanged() {
	for _, cb := range g.networkChangedCallbacks {
		cb(g)
	}
}

func (g *Network) LocalDeviceIdChanged(nodeId int64, nodeInfo *pb.NodeInfo) {
	if nodeId == 0 {
		return
	}

	if g.localDeviceId != nodeId {
		g.localDeviceId = nodeId
		if nodeId > 0 && !g.NodeIdExists(nodeId) {
			g.AddNode(NewNodeDevice(nodeId, true, "local"))
			logger.WithField("device", utils.FmtNodeId(nodeId)).Warn("Local device not found in graph, adding it. Will be an isolated node")
		}
	}

	if nodeInfo != nil {
		dev, err := g.GetNodeDevice(nodeId)
		if err != nil {
			logger.WithField("device", utils.FmtNodeId(nodeId)).Error("Failed to get local device")
			return
		}
		dev.Device().SetNodeType(NodeType(nodeInfo.NodeType))
		dev.Device().SetFriendlyName(nodeInfo.FriendlyName)
		dev.Device().SetFirmware(nodeInfo.FirmwareVersion)
		dev.Device().SetLibVersion(nodeInfo.LibVersion)
		dev.Device().SetCompileTimeString(nodeInfo.CompileTime)
	}
	g.NotifyNetworkChanged()
}

func (g *Network) LocalDeviceId() int64 {
	return g.localDeviceId
}

func (g *Network) IsLocalDevice(node graph.Node) bool {
	return g.localDeviceId == node.ID()
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
	network.networkId = g.networkId
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

func NewNetwork(localDeviceId int64, networkId int) *Network {
	network := Network{localDeviceId: localDeviceId, networkId: networkId}
	network.WeightedDirectedGraph = *simple.NewWeightedDirectedGraph(0, math.Inf(1))
	if localDeviceId > 0 {
		network.AddNode(NewNodeDevice(localDeviceId, true, "local"))
	}
	return &network
}

func NewNeworkFromFile(filename string, localDeviceId int64, networkId int) (*Network, error) {
	network := Network{localDeviceId: localDeviceId, networkId: networkId}
	network.WeightedDirectedGraph = *simple.NewWeightedDirectedGraph(0, math.Inf(1))
	err := network.readGraph(filename)
	if err != nil {
		return nil, err
	}

	if localDeviceId > 0 && !network.NodeIdExists(localDeviceId) {
		network.AddNode(NewNodeDevice(localDeviceId, true, "local"))
		logger.WithField("device", utils.FmtNodeId(localDeviceId)).Warn("Local device not found in graph, adding it. Will be an isolated node")
	}
	return &network, nil
}
