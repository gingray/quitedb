package http

import (
	"io"

	"github.com/gin-gonic/gin"
	"github.com/gingray/quitedb/internal/http/handler"
	"github.com/gingray/quitedb/internal/store"
)

type Router struct {
	db *store.Db
}

func NewRouter(db *store.Db) *Router {
	return &Router{db: db}
}

func (r *Router) SetupRoutes(router *gin.Engine) {
	probeHandler := handler.NewProbeHandler()
	router.GET("/health", probeHandler.HealthHandler)
	router.GET("/ready", probeHandler.ReadyHandler)
	router.GET("/", r.Root)
	router.GET("/get/:key", r.GetKey)
	router.POST("/put/:key", r.PutKey)

}

func (r *Router) Root(c *gin.Context) {
	c.String(200, "OK")
}

func (r *Router) GetKey(c *gin.Context) {
	key := c.Param("key")
	value := r.db.Get(key)
	c.String(200, value.(string))
}

func (r *Router) PutKey(c *gin.Context) {
	key := c.Param("key")
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithStatus(400)
		return
	}

	r.db.Put(key, string(data))
	c.Status(200)
}
