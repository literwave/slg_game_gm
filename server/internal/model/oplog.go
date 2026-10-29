package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OperationLog GM 操作审计日志，由中间件对所有写操作自动记录。
type OperationLog struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AdminID   string             `bson:"admin_id" json:"admin_id"`
	Username  string             `bson:"username" json:"username"`
	Method    string             `bson:"method" json:"method"`
	Path      string             `bson:"path" json:"path"`
	Query     string             `bson:"query,omitempty" json:"query,omitempty"`
	Body      string             `bson:"body,omitempty" json:"body,omitempty"`
	IP        string             `bson:"ip" json:"ip"`
	Status    int                `bson:"status" json:"status"`
	LatencyMs int64              `bson:"latency_ms" json:"latency_ms"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
