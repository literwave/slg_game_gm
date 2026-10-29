package errcode

import "net/http"

// Error 业务错误，携带业务码与对应的 HTTP 状态码。
type Error struct {
	Code   int
	Msg    string
	Status int
}

func (e *Error) Error() string { return e.Msg }

// WithMsg 复制一个错误并替换提示信息。
func (e *Error) WithMsg(msg string) *Error {
	return &Error{Code: e.Code, Msg: msg, Status: e.Status}
}

func New(code int, status int, msg string) *Error {
	return &Error{Code: code, Msg: msg, Status: status}
}

var (
	InvalidParams = New(10001, http.StatusBadRequest, "参数错误")
	Unauthorized  = New(10002, http.StatusUnauthorized, "未登录或登录已过期")
	Forbidden     = New(10003, http.StatusForbidden, "权限不足")
	NotFound      = New(10004, http.StatusNotFound, "资源不存在")
	Conflict      = New(10005, http.StatusConflict, "资源已存在")
	Internal      = New(10500, http.StatusInternalServerError, "服务器内部错误")

	LoginFailed  = New(20001, http.StatusUnauthorized, "用户名或密码错误")
	UserDisabled = New(20002, http.StatusForbidden, "账号已被禁用")

	GameServerUnavailable = New(30001, http.StatusBadGateway, "游戏服务器不可用")
	GameCallFailed        = New(30002, http.StatusBadGateway, "游戏服务器返回错误")
)
