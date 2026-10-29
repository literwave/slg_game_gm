package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"slg_game_gm/server/internal/model"
)

// 操作日志保留时长，超过后由 MongoDB TTL 索引自动清理。
const opLogRetention = 180 * 24 * time.Hour

type OpLogRepo struct {
	base[model.OperationLog]
}

func NewOpLogRepo(db *mongo.Database) *OpLogRepo {
	return &OpLogRepo{base[model.OperationLog]{coll: db.Collection("operation_logs")}}
}

func (r *OpLogRepo) EnsureIndexes(ctx context.Context) error {
	return r.createIndexes(ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().SetExpireAfterSeconds(int32(opLogRetention.Seconds())),
		},
		mongo.IndexModel{Keys: bson.D{{Key: "username", Value: 1}, {Key: "created_at", Value: -1}}},
	)
}

func (r *OpLogRepo) Create(ctx context.Context, l *model.OperationLog) error {
	l.ID = primitive.NewObjectID()
	return r.insert(ctx, l)
}

type OpLogFilter struct {
	Username string
	Method   string
	Path     string
	StartAt  time.Time
	EndAt    time.Time
}

func (r *OpLogRepo) List(ctx context.Context, f OpLogFilter, page, size int64) ([]model.OperationLog, int64, error) {
	filter := bson.M{}
	if f.Username != "" {
		filter["username"] = f.Username
	}
	if f.Method != "" {
		filter["method"] = f.Method
	}
	if f.Path != "" {
		filter["path"] = fuzzy(f.Path)
	}
	timeRange := bson.M{}
	if !f.StartAt.IsZero() {
		timeRange["$gte"] = f.StartAt
	}
	if !f.EndAt.IsZero() {
		timeRange["$lte"] = f.EndAt
	}
	if len(timeRange) > 0 {
		filter["created_at"] = timeRange
	}
	return r.findPage(ctx, filter, bson.D{{Key: "created_at", Value: -1}}, page, size)
}
