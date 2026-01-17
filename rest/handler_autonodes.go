package rest

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"leguru.net/m/v2/logger"
	"leguru.net/m/v2/meshmesh"
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

	params := req.toGetListParams()
	jsonNodes := h.fillNodesArrays(h.starPath.GetNetwork())
	sort.Slice(jsonNodes, func(i, j int) bool {
		return jsonNodes[i].Sort(jsonNodes[j], params.SortType, params.SortBy)
	})

	jsonNodesOut := []MeshNode{}

	if params.Offset < len(jsonNodes) {
		if params.Limit >= len(jsonNodes) {
			params.Limit = len(jsonNodes) - 1
		}
		jsonNodesOut = jsonNodes[params.Offset : params.Limit+1]
	}

	c.Header("Content-Range", fmt.Sprintf("%d-%d/%d", params.Offset, params.Limit+1, len(jsonNodes)))
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

// @Id updateAutoNode
// @Summary Update auto node
// @Tags    AutoNodes
// @Accept  json
// @Produce json
// @Param   id path string true "Auto Node ID"
// @Param   node body UpdateAutoNodeRequest true "Update auto node request"
// @Success 200 {object} MeshNode
// @Failure 400 {object} string
// @Router /api/autoNodes/{id} [put]
func (h *Handler) updateAutoNode(c *gin.Context) {
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

	network := h.starPath.GetNetwork()
	dev, err := network.GetNodeDevice(int64(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Node not found: " + err.Error()})
		return
	}

	dev.Device().SetTag(req.Tag)
	dev.Device().SetInUse(req.InUse)
	network.NotifyNetworkChanged()

	jsonNode := h.fillNodeStruct(dev, true, network)
	errors := []error{}

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
