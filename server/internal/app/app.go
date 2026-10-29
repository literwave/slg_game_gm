// Package app 负责依赖组装与服务生命周期管理。
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"slg_game_gm/server/internal/config"
	"slg_game_gm/server/internal/database"
	"slg_game_gm/server/internal/gameclient"
	"slg_game_gm/server/internal/handler"
	"slg_game_gm/server/internal/repository"
	"slg_game_gm/server/internal/router"
	"slg_game_gm/server/internal/service"
	"slg_game_gm/server/pkg/jwtutil"
)

type App struct {
	cfg    *config.Config
	mongo  *mongo.Client
	server *http.Server
}

type indexer interface {
	EnsureIndexes(ctx context.Context) error
}

func New(ctx context.Context, cfg *config.Config) (*App, error) {
	gin.SetMode(cfg.Server.Mode)

	client, db, err := database.Connect(ctx, cfg.Mongo)
	if err != nil {
		return nil, err
	}

	adminRepo := repository.NewAdminRepo(db)
	serverRepo := repository.NewGameServerRepo(db)
	mailRepo := repository.NewMailRepo(db)
	annRepo := repository.NewAnnouncementRepo(db)
	opLogRepo := repository.NewOpLogRepo(db)
	for _, r := range []indexer{adminRepo, serverRepo, mailRepo, annRepo, opLogRepo} {
		if err := r.EnsureIndexes(ctx); err != nil {
			return nil, fmt.Errorf("ensure indexes: %w", err)
		}
	}

	var game gameclient.Client
	if cfg.Game.Mock {
		slog.Warn("using mock game client")
		game = gameclient.NewMock()
	} else {
		game = gameclient.NewHTTP(cfg.Game.SignSecret, time.Duration(cfg.Game.TimeoutSeconds)*time.Second)
	}

	jwt := jwtutil.NewManager(cfg.JWT.Secret, time.Duration(cfg.JWT.ExpireHours)*time.Hour)
	adminSvc := service.NewAdminService(adminRepo, jwt)
	serverSvc := service.NewGameServerService(serverRepo, game, cfg.Game.Mock)
	playerSvc := service.NewPlayerService(serverSvc, game)
	mailSvc := service.NewMailService(mailRepo, serverSvc, game)
	annSvc := service.NewAnnouncementService(annRepo, serverSvc, game)
	opLogSvc := service.NewOpLogService(opLogRepo)
	dashboardSvc := service.NewDashboardService(serverSvc, mailRepo)

	if err := adminSvc.EnsureSuperAdmin(ctx, cfg.Admin.Username, cfg.Admin.Password); err != nil {
		return nil, fmt.Errorf("ensure super admin: %w", err)
	}

	engine := router.New(router.Deps{
		Admins:       adminSvc,
		OpLogs:       opLogSvc,
		Auth:         handler.NewAuthHandler(adminSvc),
		Admin:        handler.NewAdminHandler(adminSvc),
		GameServer:   handler.NewGameServerHandler(serverSvc),
		Player:       handler.NewPlayerHandler(playerSvc),
		Mail:         handler.NewMailHandler(mailSvc),
		Announcement: handler.NewAnnouncementHandler(annSvc),
		System:       handler.NewSystemHandler(dashboardSvc, opLogSvc),
		WebDist:      cfg.Server.WebDist,
	})

	return &App{
		cfg:   cfg,
		mongo: client,
		server: &http.Server{
			Addr:              cfg.Server.Addr,
			Handler:           engine,
			ReadHeaderTimeout: 10 * time.Second,
		},
	}, nil
}

// Run 启动 HTTP 服务，ctx 取消后优雅退出。
func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		slog.Info("gm server listening", "addr", a.cfg.Server.Addr)
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	slog.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := a.server.Shutdown(shutdownCtx); err != nil {
		slog.Error("http shutdown", "err", err)
	}
	return a.mongo.Disconnect(shutdownCtx)
}
