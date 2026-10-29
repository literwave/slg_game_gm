package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"slg_game_gm/server/internal/model"
	"slg_game_gm/server/internal/service"
	"slg_game_gm/server/pkg/errcode"
	"slg_game_gm/server/pkg/response"
)

const ctxUserKey = "gm_user"

// Auth 校验 Bearer Token，并将当前管理员写入上下文。
func Auth(admins *service.AdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if !ok || token == "" {
			response.Error(c, errcode.Unauthorized)
			return
		}
		u, err := admins.Authenticate(c.Request.Context(), token)
		if err != nil {
			response.Error(c, err)
			return
		}
		c.Set(ctxUserKey, u)
		c.Next()
	}
}

// RequireRole 要求当前用户角色不低于 min。
func RequireRole(min model.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		u := CurrentUser(c)
		if u == nil || u.Role.Level() < min.Level() {
			response.Error(c, errcode.Forbidden)
			return
		}
		c.Next()
	}
}

func CurrentUser(c *gin.Context) *model.AdminUser {
	if v, ok := c.Get(ctxUserKey); ok {
		return v.(*model.AdminUser)
	}
	return nil
}
