package handler

import (
	"github.com/gin-gonic/gin"

	"slg_game_gm/server/internal/dto"
	"slg_game_gm/server/internal/middleware"
	"slg_game_gm/server/internal/service"
	"slg_game_gm/server/pkg/response"
)

type MailHandler struct {
	mails *service.MailService
}

func NewMailHandler(mails *service.MailService) *MailHandler {
	return &MailHandler{mails: mails}
}

func (h *MailHandler) List(c *gin.Context) {
	var q dto.MailListQuery
	if !bind(c, &q) {
		return
	}
	q.Normalize()
	list, total, err := h.mails.List(c.Request.Context(), q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total)
}

func (h *MailHandler) Send(c *gin.Context) {
	var req dto.SendMailReq
	if !bind(c, &req) {
		return
	}
	mail, err := h.mails.Send(c.Request.Context(), req, middleware.CurrentUser(c).Username)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, mail)
}
