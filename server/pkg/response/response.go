package response

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"slg_game_gm/server/pkg/errcode"
)

type Body struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

type PageData struct {
	List  any   `json:"list"`
	Total int64 `json:"total"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Body{Code: 0, Msg: "ok", Data: data})
}

func Page(c *gin.Context, list any, total int64) {
	OK(c, PageData{List: list, Total: total})
}

// Error 统一错误输出：业务错误按其定义返回，未知错误记录日志后返回 500。
func Error(c *gin.Context, err error) {
	var e *errcode.Error
	if !errors.As(err, &e) {
		slog.Error("unhandled error", "path", c.FullPath(), "err", err)
		e = errcode.Internal
	}
	c.AbortWithStatusJSON(e.Status, Body{Code: e.Code, Msg: e.Msg})
}

func BadRequest(c *gin.Context, err error) {
	Error(c, errcode.InvalidParams.WithMsg("参数错误: "+err.Error()))
}
