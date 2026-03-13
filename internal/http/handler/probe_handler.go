package handler

import "github.com/gin-gonic/gin"

type ProbeHandler struct {
}

func NewProbeHandler() *ProbeHandler {
	return &ProbeHandler{}
}

func (h *ProbeHandler) ReadyHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "ready",
	})
}

func (h *ProbeHandler) HealthHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "healthy",
	})
}
