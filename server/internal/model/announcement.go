package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AnnouncementStatus string

const (
	AnnouncementDraft     AnnouncementStatus = "draft"
	AnnouncementPublished AnnouncementStatus = "published"
)

// Announcement 跑马灯/登录公告。ServerIDs 为空表示全部区服。
type Announcement struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ServerIDs       []int              `bson:"server_ids" json:"server_ids"`
	Title           string             `bson:"title" json:"title"`
	Content         string             `bson:"content" json:"content"`
	StartAt         time.Time          `bson:"start_at" json:"start_at"`
	EndAt           time.Time          `bson:"end_at" json:"end_at"`
	IntervalMinutes int                `bson:"interval_minutes" json:"interval_minutes"`
	Status          AnnouncementStatus `bson:"status" json:"status"`
	PublishedAt     *time.Time         `bson:"published_at,omitempty" json:"published_at,omitempty"`
	Results         []DispatchResult   `bson:"results,omitempty" json:"results,omitempty"`
	Operator        string             `bson:"operator" json:"operator"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}
