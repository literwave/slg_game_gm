package handler

import (
	"github.com/gin-gonic/gin"

	"slg_game_gm/server/internal/dto"
	"slg_game_gm/server/internal/service"
	"slg_game_gm/server/pkg/response"
)

type SystemHandler struct {
	dashboard *service.DashboardService
	oplogs    *service.OpLogService
}

func NewSystemHandler(dashboard *service.DashboardService, oplogs *service.OpLogService) *SystemHandler {
	return &SystemHandler{dashboard: dashboard, oplogs: oplogs}
}

func (h *SystemHandler) Dashboard(c *gin.Context) {
	stats, err := h.dashboard.Stats(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, stats)
}

func (h *SystemHandler) OpLogs(c *gin.Context) {
	var q dto.OpLogQuery
	if !bind(c, &q) {
		return
	}
	q.Normalize()
	list, total, err := h.oplogs.List(c.Request.Context(), q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total)
}
