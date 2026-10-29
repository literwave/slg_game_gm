package gameclient

import (
	"context"
	"fmt"
	"hash/fnv"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const mockPlayersPerServer = 200

var mockAlliances = []string{"", "苍狼", "龙骑", "北境", "天策", "破军"}

// mockClient 内存模拟的游戏服，便于在没有真实游戏服时开发和联调。
type mockClient struct {
	mu      sync.Mutex
	servers map[int]map[int64]*Player
}

func NewMock() Client {
	return &mockClient{servers: make(map[int]map[int64]*Player)}
}

// players 懒加载生成某区服的玩家数据，调用方需持有锁。
func (m *mockClient) players(serverID int) map[int64]*Player {
	if ps, ok := m.servers[serverID]; ok {
		return ps
	}
	rnd := rand.New(rand.NewSource(int64(serverID)))
	now := time.Now()
	ps := make(map[int64]*Player, mockPlayersPerServer)
	for i := 1; i <= mockPlayersPerServer; i++ {
		id := int64(serverID)*1_000_000 + int64(i)
		resources := make(map[string]int64, len(ResourceTypes))
		for _, r := range ResourceTypes {
			resources[r] = rnd.Int63n(5_000_000)
		}
		ps[id] = &Player{
			PlayerID:    id,
			ServerID:    serverID,
			Name:        fmt.Sprintf("领主%04d", i),
			Level:       1 + rnd.Intn(30),
			VIP:         rnd.Intn(16),
			Power:       rnd.Int63n(50_000_000),
			Alliance:    mockAlliances[rnd.Intn(len(mockAlliances))],
			Online:      rnd.Intn(3) == 0,
			Resources:   resources,
			LastLoginAt: now.Add(-time.Duration(rnd.Intn(72*3600)) * time.Second),
			CreatedAt:   now.Add(-time.Duration(rnd.Intn(90*24)) * time.Hour),
		}
	}
	m.servers[serverID] = ps
	return ps
}

func (m *mockClient) player(serverID int, id int64) (*Player, error) {
	p, ok := m.players(serverID)[id]
	if !ok {
		return nil, ErrPlayerNotFound
	}
	return p, nil
}

func clonePlayer(p *Player) Player {
	cp := *p
	cp.Resources = make(map[string]int64, len(p.Resources))
	for k, v := range p.Resources {
		cp.Resources[k] = v
	}
	return cp
}

func mockServerID(t Target) int {
	if t.ServerID > 0 {
		return t.ServerID
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(t.Addr))
	return int(h.Sum32()%9000) + 1000
}

func (m *mockClient) Status(_ context.Context, t Target) (*ServerStatus, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	sid := mockServerID(t)
	ps := m.players(sid)
	s := &ServerStatus{ServerID: sid, Registered: len(ps)}
	for _, p := range ps {
		if p.Online {
			s.Online++
		}
	}
	return s, nil
}

func (m *mockClient) SearchPlayers(_ context.Context, t Target, keyword string) ([]Player, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id, idErr := strconv.ParseInt(keyword, 10, 64)
	list := make([]Player, 0)
	for _, p := range m.players(t.ServerID) {
		match := keyword == "" || strings.Contains(p.Name, keyword) || (idErr == nil && p.PlayerID == id)
		if match {
			list = append(list, clonePlayer(p))
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].PlayerID < list[j].PlayerID })
	if len(list) > 100 {
		list = list[:100]
	}
	return list, nil
}

func (m *mockClient) GetPlayer(_ context.Context, t Target, playerID int64) (*Player, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, err := m.player(t.ServerID, playerID)
	if err != nil {
		return nil, err
	}
	cp := clonePlayer(p)
	return &cp, nil
}

func (m *mockClient) BanPlayer(_ context.Context, t Target, req BanRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, err := m.player(t.ServerID, req.PlayerID)
	if err != nil {
		return err
	}
	p.Banned, p.BanUntil, p.BanReason, p.Online = true, req.Until, req.Reason, false
	return nil
}

func (m *mockClient) UnbanPlayer(_ context.Context, t Target, playerID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, err := m.player(t.ServerID, playerID)
	if err != nil {
		return err
	}
	p.Banned, p.BanUntil, p.BanReason = false, nil, ""
	return nil
}

func (m *mockClient) KickPlayer(_ context.Context, t Target, playerID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, err := m.player(t.ServerID, playerID)
	if err != nil {
		return err
	}
	p.Online = false
	return nil
}

func (m *mockClient) ModifyResources(_ context.Context, t Target, playerID int64, changes map[string]int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, err := m.player(t.ServerID, playerID)
	if err != nil {
		return err
	}
	for k, delta := range changes {
		p.Resources[k] = max(p.Resources[k]+delta, 0)
	}
	return nil
}

func (m *mockClient) SendMail(_ context.Context, t Target, req MailRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, id := range req.PlayerIDs {
		if _, err := m.player(t.ServerID, id); err != nil {
			return &CallError{Code: codePlayerNotFound, Msg: fmt.Sprintf("玩家 %d 不存在", id)}
		}
	}
	return nil
}

func (m *mockClient) PushAnnouncement(context.Context, Target, AnnouncementRequest) error {
	return nil
}
