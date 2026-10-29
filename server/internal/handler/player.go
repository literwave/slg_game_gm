package handler

import (
	"github.com/gin-gonic/gin"

	"slg_game_gm/server/internal/dto"
	"slg_game_gm/server/internal/gameclient"
	"slg_game_gm/server/internal/service"
	"slg_game_gm/server/pkg/response"
)

type PlayerHandler struct {
	players *service.PlayerService
}

func NewPlayerHandler(players *service.PlayerService) *PlayerHandler {
	return &PlayerHandler{players: players}
}

func (h *PlayerHandler) Search(c *gin.Context) {
	var q dto.PlayerSearchQuery
	if !bind(c, &q) {
		return
	}
	list, err := h.players.Search(c.Request.Context(), q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, list)
}

func (h *PlayerHandler) Get(c *gin.Context) {
	serverID, playerID, ok := playerParams(c)
	if !ok {
		return
	}
	p, err := h.players.Get(c.Request.Context(), serverID, playerID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, p)
}

func (h *PlayerHandler) ResourceTypes(c *gin.Context) {
	response.OK(c, gameclient.ResourceTypes)
}

func (h *PlayerHandler) Ban(c *gin.Context) {
	serverID, playerID, ok := playerParams(c)
	if !ok {
		return
	}
	var req dto.BanReq
	if !bind(c, &req) {
		return
	}
	h.done(c, h.players.Ban(c.Request.Context(), serverID, playerID, req))
}

func (h *PlayerHandler) Unban(c *gin.Context) {
	serverID, playerID, ok := playerParams(c)
	if !ok {
		return
	}
	h.done(c, h.players.Unban(c.Request.Context(), serverID, playerID))
}

func (h *PlayerHandler) Kick(c *gin.Context) {
	serverID, playerID, ok := playerParams(c)
	if !ok {
		return
	}
	h.done(c, h.players.Kick(c.Request.Context(), serverID, playerID))
}

func (h *PlayerHandler) ModifyResources(c *gin.Context) {
	serverID, playerID, ok := playerParams(c)
	if !ok {
		return
	}
	var req dto.ModifyResourcesReq
	if !bind(c, &req) {
		return
	}
	h.done(c, h.players.ModifyResources(c.Request.Context(), serverID, playerID, req))
}

func (h *PlayerHandler) done(c *gin.Context, err error) {
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}
