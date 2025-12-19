package rest

import (
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"leguru.net/m/v2/logger"
	"leguru.net/m/v2/managerui"
)

/*func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		keys := make([]string, 0, len(c.Request.Header))
		for k := range c.Request.Header {
			keys = append(keys, k)
		}

		for _, k := range keys {
			logger.WithField("key", k).WithField("value", c.Request.Header[k]).Info("Request header")
		}
	}
}*/

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
	//g.Use(Logger())
	serveStaticFiles(g)

	router.Register(g)
	go g.Run(bindAddress)
}
