package rest

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vincent-petithory/dataurl"

	"leguru.net/m/v2/graph"
	"leguru.net/m/v2/logger"
	"leguru.net/m/v2/meshmesh"
	"leguru.net/m/v2/utils"
)

func (h *Handler) nodeInfoGetCmd(m *MeshNode) error {
	network := graph.GetMainNetwork()
	protocol := meshmesh.FindBestProtocol(meshmesh.MeshNodeId(m.ID), network)
	rep, err := h.serialConn.SendReceiveApiProt(meshmesh.FirmRevApiRequest{}, protocol, meshmesh.MeshNodeId(m.ID), network)
	if err != nil {
		return err
	}
	rev := rep.(meshmesh.FirmRevApiReply)

	rep, err = h.serialConn.SendReceiveApiProt(meshmesh.NodeConfigApiRequest{}, protocol, meshmesh.MeshNodeId(m.ID), network)
	if err != nil {
		return err
	}
	cfg := rep.(meshmesh.NodeConfigApiReply)

	m.Revision = utils.TruncateZeros(rev.Revision)
	m.DevTag = utils.TruncateZeros(cfg.Tag)
	m.Channel = int8(cfg.Channel)
	m.TxPower = int8(cfg.TxPower)
	m.Groups = int(cfg.Groups)
	m.Binded = int(cfg.BindedServer)
	m.Flags = int(cfg.Flags)

	return nil
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

	if h.firmwareUploadProcedure == nil || h.firmwareUploadProcedure.IsComplete() {
		if withInfo {
			err := h.nodeInfoGetCmd(&jsonNode)
			if err != nil {
				jsonNode.Error = err.Error()
			} else {
				dev.Device().SetFirmware(jsonNode.Revision)
			}
		}
	}

	return jsonNode
}

// @Id getNodes
// @Summary Get nodes
// @Tags    Nodes
// @Accept  json
// @Produce json
// @Param   login body GetListRequest true "Get list request"
// @Success 200 {array} MeshNode
// @Failure 400 {object} string
// @Router /api/nodes [get]
func (h *Handler) getNodes(c *gin.Context) {
	var req GetListRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	p := req.toGetListParams()

	network := graph.GetMainNetwork()
	nodes := network.Nodes()
	jsonNodes := make([]MeshNode, 0, nodes.Len())
	for nodes.Next() {
		dev := nodes.Node().(graph.NodeDevice)
		jsonNodes = append(jsonNodes, MeshNode{
			ID:      uint(dev.ID()),
			Tag:     string(dev.Device().Tag()),
			InUse:   dev.Device().InUse(),
			Path:    graph.FmtNodePath(network, dev),
			IsLocal: dev.ID() == network.LocalDeviceId(),
			FirmRev: dev.Device().Firmware(),
		})
	}

	sort.Slice(jsonNodes, func(i, j int) bool {
		return jsonNodes[i].Sort(jsonNodes[j], p.SortType, p.SortBy)
	})

	jsonNodesOut := []MeshNode{}

	if p.Offset < len(jsonNodes) {
		if p.Limit >= len(jsonNodes) {
			p.Limit = len(jsonNodes) - 1
		}
		jsonNodesOut = jsonNodes[p.Offset : p.Limit+1]
	}

	c.Header("Content-Range", fmt.Sprintf("%d-%d/%d", p.Offset, p.Limit+1, len(jsonNodes)))
	c.JSON(http.StatusOK, jsonNodesOut)
}

// @Id createNode
// @Summary Create node
// @Tags    Nodes
// @Accept  json
// @Produce json
// @Param   node body CreateNodeRequest true "Create node request"
// @Success 200 {object} MeshNode
// @Failure 400 {object} string
// @Router /api/nodes [post]
func (h *Handler) createNode(c *gin.Context) {
	req := CreateNodeRequest{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	network := graph.GetMainNetwork()
	_, err = network.GetNodeDevice(int64(req.ID))
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Node already exists"})
		return
	}

	dev := graph.NewNodeDevice(int64(req.ID), req.InUse, req.Tag)
	network.AddNode(dev)
	network.NotifyNetworkChanged()

	jsonNode := h.fillNodeStruct(dev, false, network)

	c.JSON(http.StatusOK, jsonNode)
}

// @Id      getOneNode
// @Summary Get one node
// @Tags    Nodes
// @Accept  json
// @Produce json
// @Param   id    path     string   true "Node ID"
// @Param   range query    string   true "Result range"
// @Param   sort  query    string   true "Sort by and type"
// @Success 200   {object} MeshNode
// @Failure 400   {string} string
// @Router  /api/nodes/{id} [get]
func (h *Handler) getOneNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	network := graph.GetMainNetwork()
	dev, err := network.GetNodeDevice(int64(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Node not found: " + err.Error()})
		return
	}

	jsonNode := h.fillNodeStruct(dev, true, network)
	c.JSON(http.StatusOK, jsonNode)
}

// @Id updateNode
// @Summary Update node
// @Tags    Nodes
// @Accept  json
// @Produce json
// @Param   id path string true "Node ID"
// @Param   node body UpdateNodeRequest true "Update node request"
// @Success 200 {object} MeshNode
// @Failure 400 {object} string
// @Router /api/nodes/{id} [put]
func (h *Handler) updateNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	req := UpdateNodeRequest{}
	err = c.ShouldBindJSON(&req)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	network := graph.GetMainNetwork()
	dev, err := network.GetNodeDevice(int64(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Node not found: " + err.Error()})
		return
	}

	dev.Device().SetTag(req.Tag)
	dev.Device().SetInUse(req.InUse)
	network.NotifyNetworkChanged()

	if req.Firmware != "" {
		firmware, err := dataurl.DecodeString(req.Firmware)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid firmware: " + err.Error()})
			return
		}

		if len(firmware.Data) > 0 {
			logger.Log().WithField("firmware", firmware.MediaType).Info("Firmware")
			err = h.uploadFirmware(int64(dev.ID()), firmware.Data)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to upload firmware: " + err.Error()})
				return
			}
		}
	}

	jsonNode := h.fillNodeStruct(dev, true, network)
	errors := []error{}

	if req.DevTag != jsonNode.DevTag {
		protocol := meshmesh.FindBestProtocol(meshmesh.MeshNodeId(dev.ID()), network)
		_, err := h.serialConn.SendReceiveApiProt(meshmesh.NodeSetTagApiRequest{Tag: req.DevTag}, protocol, meshmesh.MeshNodeId(dev.ID()), network)
		if err != nil {
			errors = append(errors, err)
		} else {
			jsonNode.DevTag = req.DevTag
		}
	}

	if req.Channel != (int8)(jsonNode.Channel) {
		protocol := meshmesh.FindBestProtocol(meshmesh.MeshNodeId(dev.ID()), network)
		_, err := h.serialConn.SendReceiveApiProt(meshmesh.NodeSetChannelApiRequest{Channel: uint8(req.Channel)}, protocol, meshmesh.MeshNodeId(dev.ID()), network)
		if err != nil {
			errors = append(errors, err)
		} else {
			jsonNode.Channel = req.Channel
		}
	}

	logger.Log().WithField("errors", errors).Info("Node update errors")
	c.JSON(http.StatusOK, jsonNode)
}

// @Id deleteNode
// @Summary Delete node
// @Tags    Nodes
// @Accept  json
// @Produce json
// @Param   id path string true "Node ID"
// @Success 200 {object} MeshNode
// @Failure 400 {object} string
// @Router /api/nodes/{id} [delete]
func (h *Handler) deleteNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	network := graph.GetMainNetwork()
	dev, err := network.GetNodeDevice(int64(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Node not found: " + err.Error()})
		return
	}

	jsonNode := h.fillNodeStruct(dev, false, network)

	network.RemoveNode(int64(id))
	network.NotifyNetworkChanged()

	c.JSON(http.StatusOK, jsonNode)
}
