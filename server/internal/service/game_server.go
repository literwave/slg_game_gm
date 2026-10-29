package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"slg_game_gm/server/internal/dto"
	"slg_game_gm/server/internal/gameclient"
	"slg_game_gm/server/internal/model"
	"slg_game_gm/server/internal/repository"
	"slg_game_gm/server/pkg/errcode"
)

// 并发向游戏服请求时，单个区服的最长等待时间。
const statusProbeTimeout = 2 * time.Second

type GameServerService struct {
	repo *repository.GameServerRepo
	game gameclient.Client
	mock bool
}

func NewGameServerService(repo *repository.GameServerRepo, game gameclient.Client, mock bool) *GameServerService {
	return &GameServerService{repo: repo, game: game, mock: mock}
}

func targetOf(s *model.GameServer) gameclient.Target {
	return gameclient.Target{ServerID: s.ServerID, Addr: s.GMAddr}
}

func (s *GameServerService) pullStatus(ctx context.Context, t gameclient.Target) (*gameclient.ServerStatus, error) {
	pctx, cancel := context.WithTimeout(ctx, statusProbeTimeout)
	defer cancel()
	return s.game.Status(pctx, t)
}

func (s *GameServerService) viewOf(gs *model.GameServer) dto.GameServerView {
	return dto.GameServerView{GameServer: *gs, Mock: s.mock}
}

// List 返回已保存的区服数据，不探测游戏服。
func (s *GameServerService) List(ctx context.Context) ([]dto.GameServerView, error) {
	servers, err := s.repo.ListAll(ctx, nil)
	if err != nil {
		return nil, err
	}
	views := make([]dto.GameServerView, len(servers))
	for i := range servers {
		views[i] = s.viewOf(&servers[i])
	}
	return views, nil
}

// Probe 探测单个区服，写入连通、在线人数，并从游戏服回写区服号。
func (s *GameServerService) Probe(ctx context.Context, id primitive.ObjectID) (*dto.GameServerView, error) {
	gs, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, repoErr(err)
	}
	now := time.Now()
	view := s.viewOf(gs)
	st, err := s.pullStatus(ctx, targetOf(gs))
	set := bson.M{"probed_at": now, "updated_at": now}
	if err != nil {
		set["reachable"] = false
		set["online"] = 0
		view.Reachable, view.Online, view.ProbedAt = false, 0, &now
	} else {
		set["reachable"] = true
		set["online"] = st.Online
		view.Reachable, view.Online, view.ProbedAt = true, st.Online, &now
		if st.ServerID > 0 {
			set["server_id"] = st.ServerID
			view.ServerID = st.ServerID
		}
	}
	if err := s.repo.Update(ctx, id, set); errors.Is(err, repository.ErrDuplicate) {
		delete(set, "server_id")
		view.ServerID = gs.ServerID
		if err := s.repo.Update(ctx, id, set); err != nil {
			return nil, repoErr(err)
		}
	} else if err != nil {
		return nil, repoErr(err)
	}
	return &view, nil
}

func (s *GameServerService) Create(ctx context.Context, req dto.GameServerReq) (*model.GameServer, error) {
	now := time.Now()
	gs := &model.GameServer{
		Name: req.Name, GMAddr: req.GMAddr, Status: req.Status,
		OpenAt: req.OpenAt, Remark: req.Remark, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.repo.Create(ctx, gs); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, errcode.Conflict.WithMsg("该区服号已接入")
		}
		return nil, err
	}
	return gs, nil
}

func (s *GameServerService) Update(ctx context.Context, id primitive.ObjectID, req dto.GameServerReq) error {
	gs, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return repoErr(err)
	}
	set := bson.M{
		"name": req.Name, "gm_addr": req.GMAddr, "status": req.Status,
		"open_at": req.OpenAt, "remark": req.Remark, "updated_at": time.Now(),
	}
	if gs.GMAddr != req.GMAddr {
		set["reachable"] = false
		set["online"] = 0
		set["probed_at"] = nil
	}
	err = s.repo.Update(ctx, id, set)
	if errors.Is(err, repository.ErrDuplicate) {
		return errcode.Conflict.WithMsg("该区服号已接入")
	}
	return repoErr(err)
}

func (s *GameServerService) Delete(ctx context.Context, id primitive.ObjectID) error {
	return repoErr(s.repo.Delete(ctx, id))
}

// Target 按区服号查找调用目标。
func (s *GameServerService) Target(ctx context.Context, serverID int) (gameclient.Target, error) {
	gs, err := s.repo.FindByServerID(ctx, serverID)
	if errors.Is(err, repository.ErrNotFound) {
		return gameclient.Target{}, errcode.NotFound.WithMsg("区服不存在")
	}
	if err != nil {
		return gameclient.Target{}, err
	}
	return targetOf(gs), nil
}

// Resolve 解析一组区服号，为空表示所有未关闭的区服。
func (s *GameServerService) Resolve(ctx context.Context, serverIDs []int) ([]model.GameServer, error) {
	servers, err := s.repo.ListAll(ctx, serverIDs)
	if err != nil {
		return nil, err
	}
	if len(serverIDs) > 0 && len(servers) != len(serverIDs) {
		return nil, errcode.InvalidParams.WithMsg("存在无效的区服号")
	}
	active := servers[:0]
	for _, gs := range servers {
		if gs.Status != model.ServerClosed {
			active = append(active, gs)
		}
	}
	if len(active) == 0 {
		return nil, errcode.InvalidParams.WithMsg("没有可用的区服")
	}
	return active, nil
}

// Broadcast 并发对多个区服执行同一操作，并收集每个区服的结果。
func (s *GameServerService) Broadcast(ctx context.Context, servers []model.GameServer,
	fn func(ctx context.Context, t gameclient.Target) error) []model.DispatchResult {
	results := make([]model.DispatchResult, len(servers))
	var wg sync.WaitGroup
	for i := range servers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			res := model.DispatchResult{ServerID: servers[i].ServerID, Success: true}
			if err := fn(ctx, targetOf(&servers[i])); err != nil {
				res.Success, res.Error = false, err.Error()
			}
			results[i] = res
		}(i)
	}
	wg.Wait()
	return results
}
