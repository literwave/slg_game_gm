package handler

import (
	"github.com/gin-gonic/gin"

	"slg_game_gm/server/internal/dto"
	"slg_game_gm/server/internal/service"
	"slg_game_gm/server/pkg/response"
)

type GameServerHandler struct {
	servers *service.GameServerService
}

func NewGameServerHandler(servers *service.GameServerService) *GameServerHandler {
	return &GameServerHandler{servers: servers}
}

// List GET /servers?status=1 时附带实时在线人数。
func (h *GameServerHandler) List(c *gin.Context) {
	list, err := h.servers.List(c.Request.Context(), c.Query("status") == "1")
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, list)
}

func (h *GameServerHandler) Create(c *gin.Context) {
	var req dto.GameServerReq
	if !bind(c, &req) {
		return
	}
	gs, err := h.servers.Create(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gs)
}

func (h *GameServerHandler) Update(c *gin.Context) {
	id, ok := objectIDParam(c, "id")
	if !ok {
		return
	}
	var req dto.GameServerReq
	if !bind(c, &req) {
		return
	}
	if err := h.servers.Update(c.Request.Context(), id, req); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *GameServerHandler) Delete(c *gin.Context) {
	id, ok := objectIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.servers.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}
