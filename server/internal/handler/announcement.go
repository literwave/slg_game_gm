package handler

import (
	"github.com/gin-gonic/gin"

	"slg_game_gm/server/internal/dto"
	"slg_game_gm/server/internal/middleware"
	"slg_game_gm/server/internal/service"
	"slg_game_gm/server/pkg/response"
)

type AnnouncementHandler struct {
	announcements *service.AnnouncementService
}

func NewAnnouncementHandler(announcements *service.AnnouncementService) *AnnouncementHandler {
	return &AnnouncementHandler{announcements: announcements}
}

func (h *AnnouncementHandler) List(c *gin.Context) {
	var q dto.AnnouncementListQuery
	if !bind(c, &q) {
		return
	}
	q.Normalize()
	list, total, err := h.announcements.List(c.Request.Context(), q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total)
}

func (h *AnnouncementHandler) Create(c *gin.Context) {
	var req dto.AnnouncementReq
	if !bind(c, &req) {
		return
	}
	a, err := h.announcements.Create(c.Request.Context(), req, middleware.CurrentUser(c).Username)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, a)
}

func (h *AnnouncementHandler) Update(c *gin.Context) {
	id, ok := objectIDParam(c, "id")
	if !ok {
		return
	}
	var req dto.AnnouncementReq
	if !bind(c, &req) {
		return
	}
	if err := h.announcements.Update(c.Request.Context(), id, req); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *AnnouncementHandler) Delete(c *gin.Context) {
	id, ok := objectIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.announcements.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *AnnouncementHandler) Publish(c *gin.Context) {
	id, ok := objectIDParam(c, "id")
	if !ok {
		return
	}
	a, err := h.announcements.Publish(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, a)
}
