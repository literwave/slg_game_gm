package service

import (
	"context"
	"slices"
	"time"

	"slg_game_gm/server/internal/dto"
	"slg_game_gm/server/internal/gameclient"
	"slg_game_gm/server/pkg/errcode"
)

type PlayerService struct {
	servers *GameServerService
	game    gameclient.Client
}

func NewPlayerService(servers *GameServerService, game gameclient.Client) *PlayerService {
	return &PlayerService{servers: servers, game: game}
}

func (s *PlayerService) Search(ctx context.Context, q dto.PlayerSearchQuery) ([]gameclient.Player, error) {
	t, err := s.servers.Target(ctx, q.ServerID)
	if err != nil {
		return nil, err
	}
	list, err := s.game.SearchPlayers(ctx, t, q.Keyword)
	return list, gameErr(err)
}

func (s *PlayerService) Get(ctx context.Context, serverID int, playerID int64) (*gameclient.Player, error) {
	t, err := s.servers.Target(ctx, serverID)
	if err != nil {
		return nil, err
	}
	p, err := s.game.GetPlayer(ctx, t, playerID)
	return p, gameErr(err)
}

func (s *PlayerService) Ban(ctx context.Context, serverID int, playerID int64, req dto.BanReq) error {
	t, err := s.servers.Target(ctx, serverID)
	if err != nil {
		return err
	}
	ban := gameclient.BanRequest{PlayerID: playerID, Reason: req.Reason}
	if req.DurationHours > 0 {
		until := time.Now().Add(time.Duration(req.DurationHours) * time.Hour)
		ban.Until = &until
	}
	return gameErr(s.game.BanPlayer(ctx, t, ban))
}

func (s *PlayerService) Unban(ctx context.Context, serverID int, playerID int64) error {
	t, err := s.servers.Target(ctx, serverID)
	if err != nil {
		return err
	}
	return gameErr(s.game.UnbanPlayer(ctx, t, playerID))
}

func (s *PlayerService) Kick(ctx context.Context, serverID int, playerID int64) error {
	t, err := s.servers.Target(ctx, serverID)
	if err != nil {
		return err
	}
	return gameErr(s.game.KickPlayer(ctx, t, playerID))
}

func (s *PlayerService) ModifyResources(ctx context.Context, serverID int, playerID int64, req dto.ModifyResourcesReq) error {
	for k := range req.Changes {
		if !slices.Contains(gameclient.ResourceTypes, k) {
			return errcode.InvalidParams.WithMsg("未知资源类型: " + k)
		}
	}
	t, err := s.servers.Target(ctx, serverID)
	if err != nil {
		return err
	}
	return gameErr(s.game.ModifyResources(ctx, t, playerID, req.Changes))
}
