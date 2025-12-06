package rest

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "leguru.net/m/v2/docs"
)

type Router interface {
	Register(g gin.IRouter)
}

type router struct {
	handler *Handler
}

func NewRouter(handler *Handler) Router {
	return &router{handler: handler}
}

func (s router) Register(g gin.IRouter) {
	h := s.handler

	g.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "FETCH"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	g.GET("/", routeFrontend)

	// use ginSwagger middleware to serve the API docs
	g.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r := g.Group("/api/v1")
	nodesGroup := r.Group("/nodes")
	{
		nodesGroup.GET("", h.getNodes)
		nodesGroup.GET("/:id", h.getOneNode)
		nodesGroup.POST("", h.createNode)
		nodesGroup.PUT("/:id", h.updateNode)
		nodesGroup.DELETE("/:id", h.deleteNode)
	}

	linksGroup := r.Group("/links")
	{
		linksGroup.GET("", h.getLinks)
		linksGroup.GET("/:id", h.getOneLink)
		linksGroup.POST("", h.createLink)
		linksGroup.PUT("/:id", h.updateLink)
		linksGroup.DELETE("/:id", h.deleteLink)
	}

	autoNodesGroup := r.Group("/autoNodes")
	{
		autoNodesGroup.GET("", h.getAutoNodes)
		autoNodesGroup.GET("/:id", h.getOneAutoNode)
	}

	autoLinksGroup := r.Group("/autoLinks")
	{
		autoLinksGroup.GET("", h.getAutoLinks)
	}

	neighborsGroup := r.Group("/neighbors")
	{
		neighborsGroup.GET("", h.getNeighbors)
		neighborsGroup.GET("/discovery/:id", h.getDiscoveryProcedureState)
		neighborsGroup.POST("/discovery", h.ctrlDiscoveryProcedure)
	}

	esphomeServersGroup := r.Group("/esphomeServers")
	{
		esphomeServersGroup.GET("", h.getEsphomeServers)
		esphomeServersGroup.GET("/connections", h.getEsphomeConnections)
	}

	esphomeConnectionsGroup := r.Group("/esphomeConnections")
	{
		esphomeConnectionsGroup.GET("", h.getEsphomeConnections)
	}
}
