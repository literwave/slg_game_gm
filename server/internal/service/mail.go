package service

import (
	"context"
	"time"

	"slg_game_gm/server/internal/dto"
	"slg_game_gm/server/internal/gameclient"
	"slg_game_gm/server/internal/model"
	"slg_game_gm/server/internal/repository"
	"slg_game_gm/server/pkg/errcode"
)

type MailService struct {
	repo    *repository.MailRepo
	servers *GameServerService
	game    gameclient.Client
}

func NewMailService(repo *repository.MailRepo, servers *GameServerService, game gameclient.Client) *MailService {
	return &MailService{repo: repo, servers: servers, game: game}
}

// Send 下发邮件到游戏服并记录。部分区服失败不会中断，结果记录在 Results 中。
func (s *MailService) Send(ctx context.Context, req dto.SendMailReq, operator string) (*model.Mail, error) {
	if req.Target == model.MailToPlayer {
		if len(req.ServerIDs) != 1 || len(req.PlayerIDs) == 0 {
			return nil, errcode.InvalidParams.WithMsg("指定玩家邮件需选择一个区服并填写玩家ID")
		}
	} else {
		req.PlayerIDs = nil
	}
	if req.Attachments == nil {
		req.Attachments = []model.Attachment{}
	}

	servers, err := s.servers.Resolve(ctx, req.ServerIDs)
	if err != nil {
		return nil, err
	}
	payload := gameclient.MailRequest{
		PlayerIDs: req.PlayerIDs, Title: req.Title, Content: req.Content,
		Attachments: req.Attachments, ExpireDays: req.ExpireDays,
	}
	results := s.servers.Broadcast(ctx, servers, func(ctx context.Context, t gameclient.Target) error {
		return gameErr(s.game.SendMail(ctx, t, payload))
	})

	serverIDs := make([]int, len(servers))
	for i, gs := range servers {
		serverIDs[i] = gs.ServerID
	}
	mail := &model.Mail{
		Target: req.Target, ServerIDs: serverIDs, PlayerIDs: req.PlayerIDs,
		Title: req.Title, Content: req.Content, Attachments: req.Attachments, ExpireDays: req.ExpireDays,
		Results: results, Operator: operator, CreatedAt: time.Now(),
	}
	if err := s.repo.Create(ctx, mail); err != nil {
		return nil, err
	}
	return mail, nil
}

func (s *MailService) List(ctx context.Context, q dto.MailListQuery) ([]model.Mail, int64, error) {
	return s.repo.List(ctx, q.ServerID, q.Keyword, q.Page, q.PageSize)
}
