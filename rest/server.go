package rest

import (
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"leguru.net/m/v2/logger"
	"leguru.net/m/v2/managerui"
)

func serveStaticFiles(g *gin.Engine) {
	//g.StaticFS("/manager", http.FS(managerui.Assets))

	fs, err := static.EmbedFolder(managerui.Assets, "data")
	if err != nil {
		logger.WithError(err).Error("Failed to embed folder")
	} else {
		g.Use(static.Serve("/", fs))
	}
}

func StartRestServer(router Router, bindAddress string) {
	var g *gin.Engine
	if logger.IsTrace() {
		g = gin.Default()
	} else {
		g = gin.New()
	}

	gin.SetMode(gin.ReleaseMode)
	serveStaticFiles(g)

	router.Register(g)
	go g.Run(bindAddress)
}
