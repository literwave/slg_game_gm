package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"slg_game_gm/server/internal/model"
)

type MailRepo struct {
	base[model.Mail]
}

func NewMailRepo(db *mongo.Database) *MailRepo {
	return &MailRepo{base[model.Mail]{coll: db.Collection("mails")}}
}

func (r *MailRepo) EnsureIndexes(ctx context.Context) error {
	return r.createIndexes(ctx,
		mongo.IndexModel{Keys: bson.D{{Key: "created_at", Value: -1}}},
		mongo.IndexModel{Keys: bson.D{{Key: "server_ids", Value: 1}, {Key: "created_at", Value: -1}}},
	)
}

func (r *MailRepo) Create(ctx context.Context, m *model.Mail) error {
	m.ID = primitive.NewObjectID()
	return r.insert(ctx, m)
}

func (r *MailRepo) List(ctx context.Context, serverID int, keyword string, page, size int64) ([]model.Mail, int64, error) {
	filter := bson.M{}
	if serverID > 0 {
		filter["server_ids"] = serverID
	}
	if keyword != "" {
		filter["title"] = fuzzy(keyword)
	}
	return r.findPage(ctx, filter, bson.D{{Key: "created_at", Value: -1}}, page, size)
}

func (r *MailRepo) CountSince(ctx context.Context, since time.Time) (int64, error) {
	return r.coll.CountDocuments(ctx, bson.M{"created_at": bson.M{"$gte": since}})
}
