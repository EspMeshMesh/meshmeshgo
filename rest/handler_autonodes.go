package rest

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"leguru.net/m/v2/graph"
)

// @Id getNodes
// @Summary Get nodes
// @Tags    Nodes
// @Accept  json
// @Produce json
// @Param   login body GetListRequest true "Get list request"
// @Success 200 {array} MeshNode
// @Failure 400 {object} string
// @Router /api/nodes [get]
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
		jsonNodes = append(jsonNodes, MeshNode{
			ID:          uint(dev.ID()),
			Tag:         string(dev.Device().Tag()),
			InUse:       dev.Device().InUse(),
			Path:        graph.FmtNodePath(network, dev),
			IsLocal:     dev.ID() == network.LocalDeviceId(),
			FirmRev:     dev.Device().Firmware(),
			compileTime: dev.Device().CompileTime(),
			CompileTime: dev.Device().CompileTimeString(),
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
