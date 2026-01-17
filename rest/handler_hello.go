package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HelloResponse struct {
	ProgramName        string `json:"program_name"`
	ProgramDescription string `json:"program_description"`
	ProgramRevision    string `json:"program_revision"`
}

var helloResponseData = HelloResponse{
	ProgramName:        "",
	ProgramDescription: "",
	ProgramRevision:    "",
}

func SetHelloResponseData(programName string, programDescription string, programRevision string) {
	helloResponseData = HelloResponse{
		ProgramName:        programName,
		ProgramDescription: programDescription,
		ProgramRevision:    programRevision,
	}
}

func (h *Handler) getHello(c *gin.Context) {
	c.JSON(http.StatusOK, helloResponseData)
}
