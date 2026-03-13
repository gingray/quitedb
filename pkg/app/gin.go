package app

import (
	"github.com/gin-gonic/gin"
)

func (a *App) WithHTTPRouter() error {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery(), gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		a.Logger.Info("GIN request", "method", param.Method, "path", param.Path, "code", param.StatusCode,
			"latency", param.Latency.String(), "ip", param.ClientIP, "error", param.ErrorMessage)
		return ""
	}))
	a.HttpRouter = router
	return nil
}
