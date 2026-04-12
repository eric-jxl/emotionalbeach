package server

import (
	"emotionalBeach/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// healthCheck godoc
// @Summary      健康检查
// @Tags         系统
// @Produce      json
// @Success      200 {object} service.HealthReport "服务健康"
// @Success      503 {object} service.HealthReport "服务降级"
// @Router       /health [get]
func healthCheck(c *gin.Context) {
	report := svc.HealthCheck()
	code := http.StatusOK
	if report.Status != "ok" {
		code = http.StatusServiceUnavailable
		zap.S().Warnw("health check degraded", "components", report.Components)
	}
	c.JSON(code, report)
}

// keep the type accessible for swagger doc references
var _ service.HealthReport

