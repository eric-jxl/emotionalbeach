// Package healthhandler wires the HTTP presentation layer for health checks.
package healthhandler

import (
	"emotionalBeach/internal/service/health"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler holds the injected health service.
type Handler struct {
	svc *healthsvc.Svc
}

// NewHandler constructs a Handler.
func NewHandler(svc *healthsvc.Svc) *Handler {
	return &Handler{svc: svc}
}

// HealthCheck godoc
// @Summary      健康检查
// @Description  返回服务整体状态、各子系统（DB / Redis）连通性及运行时指标
// @Tags         系统
// @Produce      json
// @Success      200 {object} healthsvc.Report "服务健康"
// @Success      503 {object} healthsvc.Report "服务降级"
// @Router       /health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	report := h.svc.Check()
	code := http.StatusOK
	if report.Status != "ok" {
		code = http.StatusServiceUnavailable
		zap.S().Warnw("health check degraded", "components", report.Components)
	}
	c.JSON(code, report)
}

