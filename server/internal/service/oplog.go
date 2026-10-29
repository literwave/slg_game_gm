package service

import (
	"context"
	"time"

	"slg_game_gm/server/internal/dto"
	"slg_game_gm/server/internal/model"
	"slg_game_gm/server/internal/repository"
)

type OpLogService struct {
	repo *repository.OpLogRepo
}

func NewOpLogService(repo *repository.OpLogRepo) *OpLogService {
	return &OpLogService{repo: repo}
}

func (s *OpLogService) Record(ctx context.Context, l *model.OperationLog) error {
	return s.repo.Create(ctx, l)
}

func (s *OpLogService) List(ctx context.Context, q dto.OpLogQuery) ([]model.OperationLog, int64, error) {
	f := repository.OpLogFilter{Username: q.Username, Method: q.Method, Path: q.Path}
	if q.StartAt > 0 {
		f.StartAt = time.UnixMilli(q.StartAt)
	}
	if q.EndAt > 0 {
		f.EndAt = time.UnixMilli(q.EndAt)
	}
	return s.repo.List(ctx, f, q.Page, q.PageSize)
}
