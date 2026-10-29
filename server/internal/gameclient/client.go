// Package gameclient 封装 GM 后台与游戏服之间的通信。
package gameclient

import (
	"context"
	"errors"
	"fmt"
	"time"

	"slg_game_gm/server/internal/model"
)

var (
	ErrUnavailable    = errors.New("game server unavailable")
	ErrPlayerNotFound = errors.New("player not found")
)

// CallError 游戏服返回的业务错误。
type CallError struct {
	Code int
	Msg  string
}

func (e *CallError) Error() string { return fmt.Sprintf("game error %d: %s", e.Code, e.Msg) }

// Target 一次调用的目标区服。
type Target struct {
	ServerID int
	Addr     string
}

// 资源类型，SLG 常见五种资源。
var ResourceTypes = []string{"gold", "food", "wood", "stone", "iron"}

type Player struct {
	PlayerID    int64            `json:"player_id"`
	ServerID    int              `json:"server_id"`
	Name        string           `json:"name"`
	Level       int              `json:"level"`
	VIP         int              `json:"vip"`
	Power       int64            `json:"power"`
	Alliance    string           `json:"alliance"`
	Online      bool             `json:"online"`
	Banned      bool             `json:"banned"`
	BanUntil    *time.Time       `json:"ban_until,omitempty"`
	BanReason   string           `json:"ban_reason,omitempty"`
	Resources   map[string]int64 `json:"resources"`
	LastLoginAt time.Time        `json:"last_login_at"`
	CreatedAt   time.Time        `json:"created_at"`
}

type ServerStatus struct {
	ServerID   int `json:"server_id"`
	Online     int `json:"online"`
	Registered int `json:"registered"`
}

type BanRequest struct {
	PlayerID int64      `json:"player_id"`
	Until    *time.Time `json:"until,omitempty"` // nil 表示永久
	Reason   string     `json:"reason"`
}

type MailRequest struct {
	PlayerIDs   []int64            `json:"player_ids,omitempty"` // 为空表示全服
	Title       string             `json:"title"`
	Content     string             `json:"content"`
	Attachments []model.Attachment `json:"attachments"`
	ExpireDays  int                `json:"expire_days"`
}

type AnnouncementRequest struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Content         string    `json:"content"`
	StartAt         time.Time `json:"start_at"`
	EndAt           time.Time `json:"end_at"`
	IntervalMinutes int       `json:"interval_minutes"`
}

type Client interface {
	Status(ctx context.Context, t Target) (*ServerStatus, error)
	SearchPlayers(ctx context.Context, t Target, keyword string) ([]Player, error)
	GetPlayer(ctx context.Context, t Target, playerID int64) (*Player, error)
	BanPlayer(ctx context.Context, t Target, req BanRequest) error
	UnbanPlayer(ctx context.Context, t Target, playerID int64) error
	KickPlayer(ctx context.Context, t Target, playerID int64) error
	ModifyResources(ctx context.Context, t Target, playerID int64, changes map[string]int64) error
	SendMail(ctx context.Context, t Target, req MailRequest) error
	PushAnnouncement(ctx context.Context, t Target, req AnnouncementRequest) error
}
