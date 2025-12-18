package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"leguru.net/m/v2/logger"
	"leguru.net/m/v2/managerui"
)

func serveStaticFiles(g *gin.Engine) {
	g.StaticFS("/manager", http.FS(managerui.Assets))
}

func StartRestServer(router Router, bindAddress string) {
	var g *gin.Engine
	if logger.IsTrace() {
		g = gin.Default()
	} else {
		g = gin.New()
	}
	gin.SetMode(gin.ReleaseMode)
	g.NoRoute(func(c *gin.Context) {
		c.Data(http.StatusNotFound, "text/html; charset=utf-8", []byte("<a href='/manager'>Go to manager</a>"))
	})
	serveStaticFiles(g)

	router.Register(g)
	go g.Run(bindAddress)
}
