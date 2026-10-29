package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"slg_game_gm/server/internal/model"
)

type AdminRepo struct {
	base[model.AdminUser]
}

func NewAdminRepo(db *mongo.Database) *AdminRepo {
	return &AdminRepo{base[model.AdminUser]{coll: db.Collection("admin_users")}}
}

func (r *AdminRepo) EnsureIndexes(ctx context.Context) error {
	return r.createIndexes(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
}

func (r *AdminRepo) Create(ctx context.Context, u *model.AdminUser) error {
	u.ID = primitive.NewObjectID()
	return r.insert(ctx, u)
}

func (r *AdminRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.AdminUser, error) {
	return r.findByID(ctx, id)
}

func (r *AdminRepo) FindByUsername(ctx context.Context, username string) (*model.AdminUser, error) {
	return r.findOne(ctx, bson.M{"username": username})
}

func (r *AdminRepo) List(ctx context.Context, keyword string, page, size int64) ([]model.AdminUser, int64, error) {
	filter := bson.M{}
	if keyword != "" {
		filter["$or"] = bson.A{
			bson.M{"username": fuzzy(keyword)},
			bson.M{"nickname": fuzzy(keyword)},
		}
	}
	return r.findPage(ctx, filter, bson.D{{Key: "created_at", Value: 1}}, page, size)
}

func (r *AdminRepo) Update(ctx context.Context, id primitive.ObjectID, set bson.M) error {
	return r.updateByID(ctx, id, set)
}

func (r *AdminRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	return r.deleteByID(ctx, id)
}
