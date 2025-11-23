package rest

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"leguru.net/m/v2/graph"
	"leguru.net/m/v2/utils"
)

// @Id getAutoLinks
// @Summary Get autoLinks
// @Tags    autoLinks
// @Accept  json
// @Produce json
// @Param   login body GetListRequest true "Get list request"
// @Success 200 {array} MeshLink
// @Failure 400 {string} string
// @Router /api/autolinks [get]
func (h *Handler) getAutoLinks(c *gin.Context) {
	var req GetListRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	p := req.toGetListParams()

	filter_to, _ := utils.ParseNodeId(p.Filter["to"])
	filter_from, _ := utils.ParseNodeId(p.Filter["from"])
	filter_any := smartInteger(p.Filter["any"])

	network := graph.GetMainNetwork()
	links := network.WeightedEdges()
	jsonLinks := make([]MeshLink, 0, links.Len())
	for links.Next() {
		edge := links.WeightedEdge()
		fromID := edge.From().ID()
		toID := edge.To().ID()

		if ((filter_to != -1 && filter_to != toID) && (filter_from != -1 && filter_from != fromID)) || (filter_any != -1 && (filter_any != fromID && filter_any != toID)) {
			continue
		}

		jsonLinks = append(jsonLinks, fillLinkStruct(edge))
	}

	// Sort array base on request fields
	sort.Slice(jsonLinks, func(i, j int) bool {
		return jsonLinks[i].Sort(jsonLinks[j], p.SortType, p.SortBy)
	})

	c.Header("Content-Range", fmt.Sprintf("%d-%d/%d", 0, len(jsonLinks), len(jsonLinks)))
	c.JSON(http.StatusOK, jsonLinks)
}
