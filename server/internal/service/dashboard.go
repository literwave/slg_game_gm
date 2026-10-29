package service

import (
	"context"
	"time"

	"slg_game_gm/server/internal/dto"
	"slg_game_gm/server/internal/repository"
)

type DashboardService struct {
	servers *GameServerService
	mails   *repository.MailRepo
}

func NewDashboardService(servers *GameServerService, mails *repository.MailRepo) *DashboardService {
	return &DashboardService{servers: servers, mails: mails}
}

func (s *DashboardService) Stats(ctx context.Context) (*dto.DashboardResp, error) {
	views, err := s.servers.List(ctx, true)
	if err != nil {
		return nil, err
	}
	resp := &dto.DashboardResp{ServerTotal: len(views), Servers: views}
	for _, v := range views {
		if v.Reachable {
			resp.ServerReachable++
			resp.OnlineTotal += v.Online
		}
	}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if resp.MailToday, err = s.mails.CountSince(ctx, today); err != nil {
		return nil, err
	}
	return resp, nil
}
