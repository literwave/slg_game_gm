package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"slg_game_gm/server/internal/model"
)

type AnnouncementRepo struct {
	base[model.Announcement]
}

func NewAnnouncementRepo(db *mongo.Database) *AnnouncementRepo {
	return &AnnouncementRepo{base[model.Announcement]{coll: db.Collection("announcements")}}
}

func (r *AnnouncementRepo) EnsureIndexes(ctx context.Context) error {
	return r.createIndexes(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "status", Value: 1}, {Key: "created_at", Value: -1}},
	})
}

func (r *AnnouncementRepo) Create(ctx context.Context, a *model.Announcement) error {
	a.ID = primitive.NewObjectID()
	return r.insert(ctx, a)
}

func (r *AnnouncementRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Announcement, error) {
	return r.findByID(ctx, id)
}

func (r *AnnouncementRepo) List(ctx context.Context, status string, page, size int64) ([]model.Announcement, int64, error) {
	filter := bson.M{}
	if status != "" {
		filter["status"] = status
	}
	return r.findPage(ctx, filter, bson.D{{Key: "created_at", Value: -1}}, page, size)
}

func (r *AnnouncementRepo) Update(ctx context.Context, id primitive.ObjectID, set bson.M) error {
	return r.updateByID(ctx, id, set)
}

func (r *AnnouncementRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	return r.deleteByID(ctx, id)
}
