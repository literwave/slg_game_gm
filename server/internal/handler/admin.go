package handler

import (
	"github.com/gin-gonic/gin"

	"slg_game_gm/server/internal/dto"
	"slg_game_gm/server/internal/middleware"
	"slg_game_gm/server/internal/service"
	"slg_game_gm/server/pkg/response"
)

type AdminHandler struct {
	admins *service.AdminService
}

func NewAdminHandler(admins *service.AdminService) *AdminHandler {
	return &AdminHandler{admins: admins}
}

func (h *AdminHandler) List(c *gin.Context) {
	var q dto.AdminListQuery
	if !bind(c, &q) {
		return
	}
	q.Normalize()
	list, total, err := h.admins.List(c.Request.Context(), q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total)
}

func (h *AdminHandler) Create(c *gin.Context) {
	var req dto.CreateAdminReq
	if !bind(c, &req) {
		return
	}
	u, err := h.admins.Create(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, u)
}

func (h *AdminHandler) Update(c *gin.Context) {
	id, ok := objectIDParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateAdminReq
	if !bind(c, &req) {
		return
	}
	if err := h.admins.Update(c.Request.Context(), middleware.CurrentUser(c), id, req); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *AdminHandler) ResetPassword(c *gin.Context) {
	id, ok := objectIDParam(c, "id")
	if !ok {
		return
	}
	var req dto.ResetPasswordReq
	if !bind(c, &req) {
		return
	}
	if err := h.admins.ResetPassword(c.Request.Context(), id, req); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *AdminHandler) Delete(c *gin.Context) {
	id, ok := objectIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.admins.Delete(c.Request.Context(), middleware.CurrentUser(c), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}
