package rest

import (
	"strconv"
	"strings"

	"leguru.net/m/v2/graph"
	"leguru.net/m/v2/meshmesh"
	"leguru.net/m/v2/meshmesh/pb"
	"leguru.net/m/v2/utils"
)

func revisionToInteger(revision string) int {
	if strings.Contains(revision, ",") {
		return 0
	}
	parts := strings.Split(revision, ".")
	if len(parts) != 3 {
		return 0
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0
	}
	return major*1000000 + minor*1000 + patch
}

func (h *Handler) fillNodeStruct(dev graph.NodeDevice, withInfo bool, network *graph.Network) MeshNode {
	jsonNode := MeshNode{
		ID:          uint(dev.ID()),
		Tag:         string(dev.Device().Tag()),
		InUse:       dev.Device().InUse(),
		IsLocal:     dev.ID() == network.LocalDeviceId(),
		FirmRev:     dev.Device().Firmware(),
		LibVersion:  dev.Device().LibVersion(),
		CompileTime: formatTimeForJson(dev.Device().CompileTime()),
		LastSeen:    formatTimeForJson(dev.Device().LastSeen()),
		Path:        graph.FmtNodePath(network, dev),
	}

	if withInfo {
		err := h.nodeInfoGetCmd(network, &jsonNode)
		if err != nil {
			jsonNode.Error = err.Error()
		} else {
			dev.Device().SetFirmware(jsonNode.DevRevision)
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

	if revisionToInteger(m.FirmRev) > 1004002 {
		rep, err = h.serialConn.SendReceiveApiProt(meshmesh.ProtoNodeInfoApiRequest{}, protocol, meshmesh.MeshNodeId(m.ID), network)
		if err != nil {
			return err
		}
		nodeInfo := rep.(*pb.NodeInfo)
		//m.Tag = nodeInfo.FriendlyName
		m.CompileTime = nodeInfo.CompileTime
		m.FirmRev = nodeInfo.FirmwareVersion
		m.LibVersion = nodeInfo.LibVersion
	}

	rep, err = h.serialConn.SendReceiveApiProt(meshmesh.NodeConfigApiRequest{}, protocol, meshmesh.MeshNodeId(m.ID), network)
	if err != nil {
		return err
	}
	cfg := rep.(meshmesh.NodeConfigApiReply)

	m.DevRevision = utils.TruncateZeros(rev.Revision)
	m.DevTag = utils.TruncateZeros(cfg.Tag)
	m.Channel = int8(cfg.Channel)
	m.TxPower = int8(cfg.TxPower)
	m.Groups = int(cfg.Groups)
	m.Binded = int(cfg.BindedServer)
	m.Flags = int(cfg.Flags)

	return nil
}
