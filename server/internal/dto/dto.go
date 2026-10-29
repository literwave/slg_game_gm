// Package dto 定义 HTTP 接口的请求与响应结构。
package dto

import (
	"time"

	"slg_game_gm/server/internal/model"
)

type PageQuery struct {
	Page     int64 `form:"page"`
	PageSize int64 `form:"page_size"`
}

func (p *PageQuery) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 100 {
		p.PageSize = 20
	}
}

// ---------- 认证 ----------

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResp struct {
	Token    string           `json:"token"`
	ExpireAt time.Time        `json:"expire_at"`
	User     *model.AdminUser `json:"user"`
}

type ChangePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=64"`
}

// ---------- 管理员 ----------

type AdminListQuery struct {
	PageQuery
	Keyword string `form:"keyword"`
}

type CreateAdminReq struct {
	Username string     `json:"username" binding:"required,min=3,max=32,alphanum"`
	Password string     `json:"password" binding:"required,min=6,max=64"`
	Nickname string     `json:"nickname" binding:"max=32"`
	Role     model.Role `json:"role" binding:"required,oneof=super_admin admin operator viewer"`
}

type UpdateAdminReq struct {
	Nickname *string     `json:"nickname" binding:"omitempty,max=32"`
	Role     *model.Role `json:"role" binding:"omitempty,oneof=super_admin admin operator viewer"`
	Disabled *bool       `json:"disabled"`
}

type ResetPasswordReq struct {
	Password string `json:"password" binding:"required,min=6,max=64"`
}

// ---------- 区服 ----------

type GameServerReq struct {
	Name   string             `json:"name" binding:"required,max=32"`
	GMAddr string             `json:"gm_addr" binding:"required,url"`
	Status model.ServerStatus `json:"status" binding:"min=0,max=3"`
	OpenAt time.Time          `json:"open_at"`
	Remark string             `json:"remark" binding:"max=256"`
}

type GameServerView struct {
	model.GameServer
	Mock bool `json:"mock"`
}

// ---------- 玩家 ----------

type PlayerSearchQuery struct {
	ServerID int    `form:"server_id" binding:"required,min=1"`
	Keyword  string `form:"keyword"`
}

type BanReq struct {
	DurationHours int    `json:"duration_hours" binding:"min=0"` // 0 表示永久
	Reason        string `json:"reason" binding:"required,max=128"`
}

type ModifyResourcesReq struct {
	Changes map[string]int64 `json:"changes" binding:"required,min=1"`
}

// ---------- 邮件 ----------

type SendMailReq struct {
	Target      model.MailTarget   `json:"target" binding:"required,oneof=player server"`
	ServerIDs   []int              `json:"server_ids"`
	PlayerIDs   []int64            `json:"player_ids"`
	Title       string             `json:"title" binding:"required,max=64"`
	Content     string             `json:"content" binding:"required,max=2000"`
	Attachments []model.Attachment `json:"attachments" binding:"omitempty,max=10,dive"`
	ExpireDays  int                `json:"expire_days" binding:"min=0,max=30"`
}

type MailListQuery struct {
	PageQuery
	ServerID int    `form:"server_id"`
	Keyword  string `form:"keyword"`
}

// ---------- 公告 ----------

type AnnouncementReq struct {
	ServerIDs       []int     `json:"server_ids"`
	Title           string    `json:"title" binding:"required,max=64"`
	Content         string    `json:"content" binding:"required,max=1000"`
	StartAt         time.Time `json:"start_at"`
	EndAt           time.Time `json:"end_at"`
	IntervalMinutes int       `json:"interval_minutes" binding:"min=0"`
}

type AnnouncementListQuery struct {
	PageQuery
	Status string `form:"status"`
}

// ---------- 操作日志 ----------

type OpLogQuery struct {
	PageQuery
	Username string `form:"username"`
	Method   string `form:"method"`
	Path     string `form:"path"`
	StartAt  int64  `form:"start_at"` // 毫秒时间戳
	EndAt    int64  `form:"end_at"`
}

// ---------- 仪表盘 ----------

type DashboardResp struct {
	ServerTotal     int              `json:"server_total"`
	ServerReachable int              `json:"server_reachable"`
	OnlineTotal     int              `json:"online_total"`
	MailToday       int64            `json:"mail_today"`
	Servers         []GameServerView `json:"servers"`
}
