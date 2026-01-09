package rest

import (
	"leguru.net/m/v2/graph"
	"leguru.net/m/v2/meshmesh"
	"leguru.net/m/v2/meshmesh/pb"
	"leguru.net/m/v2/utils"
)

func (h *Handler) fillNodesArrays(network *graph.Network) []MeshNode {
	nodes := network.Nodes()
	nodesArray := make([]MeshNode, 0, nodes.Len())
	for nodes.Next() {
		dev := nodes.Node().(graph.NodeDevice)
		// Create MeshNode struct
		d := dev.Device()
		nodesArray = append(nodesArray, MeshNode{
			ID:          uint(dev.ID()),
			Tag:         string(d.Tag()),
			InUse:       d.InUse(),
			DeepSleep:   d.DeepSleep(),
			Path:        graph.FmtNodePath(network, dev),
			IsLocal:     dev.ID() == network.LocalDeviceId(),
			FirmRev:     d.Firmware(),
			LibVersion:  d.LibVersion(),
			CompileTime: formatTimeForJson(d.CompileTime()),
			LastSeen:    formatTimeForJson(d.LastSeen()),
			DevType:     d.NodeTypeString(),
			compileTime: d.CompileTime(),
			lastSeen:    d.LastSeen(),
		})
	}
	return nodesArray
}

func (h *Handler) fillNodeStruct(dev graph.NodeDevice, withInfo bool, network *graph.Network) MeshNode {

	d := dev.Device()
	jsonNode := MeshNode{
		ID:          uint(dev.ID()),
		Tag:         string(d.Tag()),
		InUse:       d.InUse(),
		DeepSleep:   d.DeepSleep(),
		IsLocal:     dev.ID() == network.LocalDeviceId(),
		FirmRev:     d.Firmware(),
		LibVersion:  d.LibVersion(),
		CompileTime: formatTimeForJson(d.CompileTime()),
		DevType:     d.NodeTypeString(),
		LastSeen:    formatTimeForJson(d.LastSeen()),
		Path:        graph.FmtNodePath(network, dev),
	}

	if withInfo {
		err := h.nodeInfoGetCmd(network, &jsonNode)
		if err != nil {
			jsonNode.Error = err.Error()
		} else {
			changed := false
			if jsonNode.DevFriendlyName != "" && jsonNode.DevFriendlyName != d.FriendlyName() {
				d.SetFriendlyName(jsonNode.DevFriendlyName)
				changed = true
			}
			if jsonNode.LibVersion != "" && jsonNode.LibVersion != d.LibVersion() {
				d.SetLibVersion(jsonNode.LibVersion)
				changed = true
			}
			if jsonNode.DevRevision != "" && jsonNode.DevRevision != d.Firmware() {
				d.SetFirmware(jsonNode.DevRevision)
				changed = true
			}
			if jsonNode.CompileTime != "" && jsonNode.CompileTime != d.CompileTimeString() {
				d.SetCompileTimeString(jsonNode.CompileTime)
				changed = true
			}
			if changed {
				network.NotifyNetworkChanged()
			}
		}
	}

	return jsonNode
}

func (h *Handler) nodeInfoGetCmd(network *graph.Network, m *MeshNode) error {
	protocol := meshmesh.FindBestProtocol(meshmesh.MeshNodeId(m.ID), network)
	rep, err := h.serialConn.SendReceiveApiProt(meshmesh.FirmRevApiRequest{}, protocol, meshmesh.MeshNodeId(m.ID), network)
	if err != nil {
		return err
	}
	rev := rep.(meshmesh.FirmRevApiReply)

	if utils.RevisionToInteger(m.FirmRev) > 1004002 {
		rep, err = h.serialConn.SendReceiveApiProt(meshmesh.ProtoNodeInfoApiRequest{}, protocol, meshmesh.MeshNodeId(m.ID), network)
		if err != nil {
			return err
		}
		nodeInfo := rep.(*pb.NodeInfo)
		m.DevFriendlyName = nodeInfo.FriendlyName
		m.CompileTime = nodeInfo.CompileTime
		m.FirmRev = nodeInfo.FirmwareVersion
		m.LibVersion = nodeInfo.LibVersion
		m.DevType = graph.EnumNodeTypeToString(graph.NodeType(nodeInfo.NodeType))
	}

	rep, err = h.serialConn.SendReceiveApiProt(meshmesh.NodeConfigApiRequest{}, protocol, meshmesh.MeshNodeId(m.ID), network)
	if err != nil {
		return err
	}
	cfg := rep.(meshmesh.NodeConfigApiReply)

	m.DevRevision = utils.TruncateZeros(rev.Revision)
	m.DevName = utils.TruncateZeros(cfg.Tag)
	m.Channel = int8(cfg.Channel)
	m.TxPower = int8(cfg.TxPower)
	m.Groups = int(cfg.Groups)
	m.Binded = int(cfg.BindedServer)
	m.Flags = int(cfg.Flags)

	return nil
}
