package rest

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"leguru.net/m/v2/graph"
	"leguru.net/m/v2/logger"
)

// @Id getAutoNodes
// @Summary Get auto formed network nodes
// @Tags    Nodes
// @Accept  json
// @Produce json
// @Param   login body GetListRequest true "Get list request"
// @Success 200 {array} MeshNode
// @Failure 400 {object} string
// @Router /api/v1/autoNodes [get]
func (h *Handler) getAutoNodes(c *gin.Context) {
	var req GetListRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	p := req.toGetListParams()

	network := h.starPath.GetNetwork()
	nodes := network.Nodes()
	jsonNodes := make([]MeshNode, 0, nodes.Len())

	for nodes.Next() {
		dev := nodes.Node().(graph.NodeDevice)
		logger.Debug("firmware name: %s", dev.Device().Firmware())
		// Create MeshNode struct
		jsonNodes = append(jsonNodes, MeshNode{
			ID:          uint(dev.ID()),
			Tag:         string(dev.Device().Tag()),
			InUse:       dev.Device().InUse(),
			Path:        graph.FmtNodePath(network, dev),
			IsLocal:     dev.ID() == network.LocalDeviceId(),
			FirmRev:     dev.Device().Firmware(),
			CompileTime: formatTimeForJson(dev.Device().CompileTime()),
			LastSeen:    formatTimeForJson(dev.Device().LastSeen()),

			compileTime: dev.Device().CompileTime(),
			lastSeen:    dev.Device().LastSeen(),
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

// @Id      getOneAutoNode
// @Summary Get one auto formed network node
// @Tags    AutoNodes
// @Accept  json
// @Produce json
// @Param   id    path     string   true "Node ID"
// @Success 200   {object} MeshNode
// @Failure 400   {string} string
// @Router  /api/v1/autoNodes/{id} [get]
func (h *Handler) getOneAutoNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	network := h.starPath.GetNetwork()
	dev, err := network.GetNodeDevice(int64(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Node not found: " + err.Error()})
		return
	}

	jsonNode := h.fillNodeStruct(dev, true, network)
	c.JSON(http.StatusOK, jsonNode)
}

// @Id deleteAutoNode
// @Summary Delete node
// @Tags    AutoNodes
// @Accept  json
// @Produce json
// @Param   id path string true "Auto Node ID"
// @Success 200 {object} MeshNode
// @Failure 400 {object} string
// @Router /api/v1/autoNodes/{id} [delete]
func (h *Handler) deleteAutoNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	network := h.starPath.GetNetwork()
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
