package rest

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"leguru.net/m/v2/meshmesh"
)

func (h *Handler) rebootNode(c *gin.Context) {
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

	protocol := meshmesh.FindBestProtocol(meshmesh.MeshNodeId(dev.ID()), network)
	_, err = h.serialConn.SendReceiveApiProt(meshmesh.NodeRebootApiRequest{Id: uint8(dev.ID())}, protocol, meshmesh.MeshNodeId(dev.ID()), network)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to reboot node: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Node rebooted"})
}
