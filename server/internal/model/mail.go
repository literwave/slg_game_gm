package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MailTarget string

const (
	MailToPlayer MailTarget = "player" // 指定玩家（单服）
	MailToServer MailTarget = "server" // 全服（可多服，空表示所有区服）
)

type Attachment struct {
	ItemID int   `bson:"item_id" json:"item_id" binding:"required,min=1"`
	Count  int64 `bson:"count" json:"count" binding:"required,min=1"`
}

// DispatchResult 下发到单个区服的结果，邮件和公告共用。
type DispatchResult struct {
	ServerID int    `bson:"server_id" json:"server_id"`
	Success  bool   `bson:"success" json:"success"`
	Error    string `bson:"error,omitempty" json:"error,omitempty"`
}

type Mail struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Target      MailTarget         `bson:"target" json:"target"`
	ServerIDs   []int              `bson:"server_ids" json:"server_ids"`
	PlayerIDs   []int64            `bson:"player_ids,omitempty" json:"player_ids,omitempty"`
	Title       string             `bson:"title" json:"title"`
	Content     string             `bson:"content" json:"content"`
	Attachments []Attachment       `bson:"attachments" json:"attachments"`
	ExpireDays  int                `bson:"expire_days" json:"expire_days"`
	Results     []DispatchResult   `bson:"results" json:"results"`
	Operator    string             `bson:"operator" json:"operator"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}
