package router

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"slg_game_gm/server/internal/handler"
	"slg_game_gm/server/internal/middleware"
	"slg_game_gm/server/internal/model"
	"slg_game_gm/server/internal/service"
	"slg_game_gm/server/pkg/errcode"
	"slg_game_gm/server/pkg/response"
)

type Deps struct {
	Admins       *service.AdminService
	OpLogs       *service.OpLogService
	Auth         *handler.AuthHandler
	Admin        *handler.AdminHandler
	GameServer   *handler.GameServerHandler
	Player       *handler.PlayerHandler
	Mail         *handler.MailHandler
	Announcement *handler.AnnouncementHandler
	System       *handler.SystemHandler
	WebDist      string
}

func New(d Deps) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Logger(), gin.Recovery(), middleware.CORS())

	r.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	api := r.Group("/api/v1")
	api.POST("/auth/login", d.Auth.Login)

	authed := api.Group("", middleware.Auth(d.Admins), middleware.OperationLog(d.OpLogs))
	operator := middleware.RequireRole(model.RoleOperator)
	admin := middleware.RequireRole(model.RoleAdmin)
	super := middleware.RequireRole(model.RoleSuperAdmin)

	authed.GET("/auth/profile", d.Auth.Profile)
	authed.PUT("/auth/password", d.Auth.ChangePassword)
	authed.GET("/dashboard", d.System.Dashboard)
	authed.GET("/meta/resource-types", d.Player.ResourceTypes)

	servers := authed.Group("/servers")
	servers.GET("", d.GameServer.List)
	servers.POST("", admin, d.GameServer.Create)
	servers.PUT("/:id", admin, d.GameServer.Update)
	servers.DELETE("/:id", admin, d.GameServer.Delete)

	players := authed.Group("/players")
	players.GET("", d.Player.Search)
	players.GET("/:server_id/:player_id", d.Player.Get)
	players.POST("/:server_id/:player_id/ban", operator, d.Player.Ban)
	players.POST("/:server_id/:player_id/unban", operator, d.Player.Unban)
	players.POST("/:server_id/:player_id/kick", operator, d.Player.Kick)
	players.POST("/:server_id/:player_id/resources", operator, d.Player.ModifyResources)

	mails := authed.Group("/mails")
	mails.GET("", d.Mail.List)
	mails.POST("", operator, d.Mail.Send)

	ann := authed.Group("/announcements")
	ann.GET("", d.Announcement.List)
	ann.POST("", operator, d.Announcement.Create)
	ann.PUT("/:id", operator, d.Announcement.Update)
	ann.DELETE("/:id", operator, d.Announcement.Delete)
	ann.POST("/:id/publish", operator, d.Announcement.Publish)

	admins := authed.Group("/admins", super)
	admins.GET("", d.Admin.List)
	admins.POST("", d.Admin.Create)
	admins.PUT("/:id", d.Admin.Update)
	admins.PUT("/:id/password", d.Admin.ResetPassword)
	admins.DELETE("/:id", d.Admin.Delete)

	authed.GET("/oplogs", admin, d.System.OpLogs)

	r.NoRoute(spaFallback(d.WebDist))
	return r
}

// spaFallback 托管前端构建产物：存在的静态文件直接返回，其余路径回退到 index.html。
func spaFallback(dist string) gin.HandlerFunc {
	enabled := false
	if dist != "" {
		if _, err := os.Stat(filepath.Join(dist, "index.html")); err == nil {
			enabled = true
		}
	}
	return func(c *gin.Context) {
		if !enabled || strings.HasPrefix(c.Request.URL.Path, "/api/") || c.Request.Method != http.MethodGet {
			response.Error(c, errcode.NotFound)
			return
		}
		file := filepath.Join(dist, filepath.FromSlash(path.Clean("/"+c.Request.URL.Path)))
		if info, err := os.Stat(file); err == nil && !info.IsDir() {
			c.File(file)
			return
		}
		c.File(filepath.Join(dist, "index.html"))
	}
}
