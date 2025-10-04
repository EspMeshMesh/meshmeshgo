package rest

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"leguru.net/m/v2/graph"
	"leguru.net/m/v2/utils"

	gr "gonum.org/v1/gonum/graph"
)

func parseFromToId(id uint) (from, to uint) {
	from = uint(id) & 0x00FFFFFF
	to = uint(id) >> 24
	return from, to
}

func linkIdFromToId(id uint) (from, to uint) {
	from = uint(id) & 0x00FFFFFF
	to = uint(id) >> 24
	return from, to
}

func fillLinkStruct(edge gr.WeightedEdge) MeshLink {
	from := edge.From().(graph.NodeDevice)
	to := edge.To().(graph.NodeDevice)

	return MeshLink{
		ID:          uint(from.ID()) + uint(to.ID())<<24,
		From:        from.ID(),
		To:          to.ID(),
		Weight:      float32(edge.Weight()),
		Description: fmt.Sprintf("from: %s to: %s", from.Device().Tag(), to.Device().Tag()),
	}
}

// @Id getLinks
// @Summary Get links
// @Tags    Links
// @Accept  json
// @Produce json
// @Param   login body GetListRequest true "Get list request"
// @Success 200 {array} MeshLink
// @Failure 400 {string} string
// @Router /api/links [get]
func (h *Handler) getLinks(c *gin.Context) {
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

// @Id getOneLink
// @Summary Get one link
// @Tags    Links
// @Accept  json
// @Produce json
// @Param   id path int true "Link ID"
// @Success 200 {object} MeshLink
// @Failure 400 {string} string
// @Router /api/links/{id} [get]
func (h *Handler) getOneLink(c *gin.Context) {
	fromToIdStr := c.Param("id")
	fromToId, err := strconv.ParseUint(fromToIdStr, 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	fromID, toID := parseFromToId(uint(fromToId))
	network := graph.GetMainNetwork()
	edge := network.WeightedEdge(int64(fromID), int64(toID))
	if edge == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Link not found"})
		return
	}

	jsonLink := fillLinkStruct(edge)
	c.JSON(http.StatusOK, jsonLink)
}

// @Id createLink
// @Summary Create link
// @Tags    Links
// @Accept  json
// @Produce json
// @Param   link body bool true "Create link request"
// @Success 200 {object} MeshLink
// @Failure 400 {object} string
// @Router /api/links [post]
func (h *Handler) createLink(c *gin.Context) {
	req := CreateLinkRequest{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	network := graph.GetMainNetwork()
	_, err = network.GetNodeDevice(req.From)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "From node not found"})
		return
	}
	_, err = network.GetNodeDevice(req.To)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "To node not found"})
		return
	}

	network.ChangeEdgeWeight(req.From, req.To, float64(req.Weight), float64(req.Weight))
	graph.NotifyMainNetworkChanged()

	edge := network.WeightedEdge(req.From, req.To)
	if edge == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Link not found"})
		return
	}

	jsonLink := fillLinkStruct(edge)
	c.JSON(http.StatusOK, jsonLink)
}

// @Id updateLink
// @Summary Update link
// @Tags    Links
// @Accept  json
// @Produce json
// @Param   id path integer true "Mixed FromTo ID"
// @Param   link body UpdateLinkRequest true "Update link request"
// @Success 200 {object} MeshLink
// @Failure 400 {object} string
// @Router /api/links/{id} [put]
func (h *Handler) updateLink(c *gin.Context) {
	fromToIdStr := c.Param("id")
	fromToId, err := strconv.ParseUint(fromToIdStr, 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	req := UpdateLinkRequest{}
	err = c.ShouldBindJSON(&req)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	fromID, toID := parseFromToId(uint(fromToId))
	network := graph.GetMainNetwork()
	edge := network.WeightedEdge(int64(fromID), int64(toID))
	if edge == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Link not found"})
		return
	}

	network.ChangeEdgeWeight(int64(fromID), int64(toID), float64(req.Weight), float64(req.Weight))
	graph.NotifyMainNetworkChanged()

	jsonLink := fillLinkStruct(edge)
	c.JSON(http.StatusOK, jsonLink)
}

// @Id deleteLink
// @Summary Delete link
// @Tags    Links
// @Accept  json
// @Produce json
// @Param   id path integer true "Mixed FromTo ID"
// @Success 200 {object} MeshLink
// @Failure 400 {string} string
// @Router /api/links/{id} [delete]
func (h *Handler) deleteLink(c *gin.Context) {
	fromToIdStr := c.Param("id")
	fromToId, err := strconv.ParseUint(fromToIdStr, 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	fromID, toID := parseFromToId(uint(fromToId))
	network := graph.GetMainNetwork()
	edge := network.WeightedEdge(int64(fromID), int64(toID))
	if edge == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Link not found"})
		return
	}

	network.RemoveEdge(int64(fromID), int64(toID))
	graph.NotifyMainNetworkChanged()

	jsonLink := fillLinkStruct(edge)
	c.JSON(http.StatusOK, jsonLink)
}
