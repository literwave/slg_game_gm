package handler

import (
	"github.com/gin-gonic/gin"

	"slg_game_gm/server/internal/dto"
	"slg_game_gm/server/internal/middleware"
	"slg_game_gm/server/internal/service"
	"slg_game_gm/server/pkg/response"
)

type AuthHandler struct {
	admins *service.AdminService
}

func NewAuthHandler(admins *service.AdminService) *AuthHandler {
	return &AuthHandler{admins: admins}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginReq
	if !bind(c, &req) {
		return
	}
	resp, err := h.admins.Login(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *AuthHandler) Profile(c *gin.Context) {
	response.OK(c, middleware.CurrentUser(c))
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordReq
	if !bind(c, &req) {
		return
	}
	if err := h.admins.ChangePassword(c.Request.Context(), middleware.CurrentUser(c), req); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}
