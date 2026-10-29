package service

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"slg_game_gm/server/internal/dto"
	"slg_game_gm/server/internal/gameclient"
	"slg_game_gm/server/internal/model"
	"slg_game_gm/server/internal/repository"
	"slg_game_gm/server/pkg/errcode"
)

type AnnouncementService struct {
	repo    *repository.AnnouncementRepo
	servers *GameServerService
	game    gameclient.Client
}

func NewAnnouncementService(repo *repository.AnnouncementRepo, servers *GameServerService, game gameclient.Client) *AnnouncementService {
	return &AnnouncementService{repo: repo, servers: servers, game: game}
}

func validateWindow(req dto.AnnouncementReq) error {
	if req.StartAt.IsZero() || !req.EndAt.After(req.StartAt) {
		return errcode.InvalidParams.WithMsg("结束时间必须晚于开始时间")
	}
	return nil
}

func (s *AnnouncementService) List(ctx context.Context, q dto.AnnouncementListQuery) ([]model.Announcement, int64, error) {
	return s.repo.List(ctx, q.Status, q.Page, q.PageSize)
}

func (s *AnnouncementService) Create(ctx context.Context, req dto.AnnouncementReq, operator string) (*model.Announcement, error) {
	if err := validateWindow(req); err != nil {
		return nil, err
	}
	now := time.Now()
	a := &model.Announcement{
		ServerIDs: req.ServerIDs, Title: req.Title, Content: req.Content,
		StartAt: req.StartAt, EndAt: req.EndAt, IntervalMinutes: req.IntervalMinutes,
		Status: model.AnnouncementDraft, Operator: operator, CreatedAt: now, UpdatedAt: now,
	}
	if a.ServerIDs == nil {
		a.ServerIDs = []int{}
	}
	if err := s.repo.Create(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

// Update 修改公告内容；已发布的公告修改后回到草稿状态，需重新发布。
func (s *AnnouncementService) Update(ctx context.Context, id primitive.ObjectID, req dto.AnnouncementReq) error {
	if err := validateWindow(req); err != nil {
		return err
	}
	if req.ServerIDs == nil {
		req.ServerIDs = []int{}
	}
	return repoErr(s.repo.Update(ctx, id, bson.M{
		"server_ids": req.ServerIDs, "title": req.Title, "content": req.Content,
		"start_at": req.StartAt, "end_at": req.EndAt, "interval_minutes": req.IntervalMinutes,
		"status": model.AnnouncementDraft, "updated_at": time.Now(),
	}))
}

func (s *AnnouncementService) Delete(ctx context.Context, id primitive.ObjectID) error {
	return repoErr(s.repo.Delete(ctx, id))
}

func (s *AnnouncementService) Publish(ctx context.Context, id primitive.ObjectID) (*model.Announcement, error) {
	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, repoErr(err)
	}
	servers, err := s.servers.Resolve(ctx, a.ServerIDs)
	if err != nil {
		return nil, err
	}
	payload := gameclient.AnnouncementRequest{
		ID: a.ID.Hex(), Title: a.Title, Content: a.Content,
		StartAt: a.StartAt, EndAt: a.EndAt, IntervalMinutes: a.IntervalMinutes,
	}
	a.Results = s.servers.Broadcast(ctx, servers, func(ctx context.Context, t gameclient.Target) error {
		return gameErr(s.game.PushAnnouncement(ctx, t, payload))
	})
	now := time.Now()
	a.Status, a.PublishedAt, a.UpdatedAt = model.AnnouncementPublished, &now, now
	err = s.repo.Update(ctx, id, bson.M{
		"status": a.Status, "published_at": now, "results": a.Results, "updated_at": now,
	})
	return a, repoErr(err)
}
