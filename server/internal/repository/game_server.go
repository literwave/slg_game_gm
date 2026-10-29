package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"slg_game_gm/server/internal/model"
)

type GameServerRepo struct {
	base[model.GameServer]
}

func NewGameServerRepo(db *mongo.Database) *GameServerRepo {
	return &GameServerRepo{base[model.GameServer]{coll: db.Collection("game_servers")}}
}

func (r *GameServerRepo) EnsureIndexes(ctx context.Context) error {
	return r.createIndexes(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "server_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
}

func (r *GameServerRepo) Create(ctx context.Context, s *model.GameServer) error {
	s.ID = primitive.NewObjectID()
	return r.insert(ctx, s)
}

func (r *GameServerRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.GameServer, error) {
	return r.findByID(ctx, id)
}

func (r *GameServerRepo) FindByServerID(ctx context.Context, serverID int) (*model.GameServer, error) {
	return r.findOne(ctx, bson.M{"server_id": serverID})
}

// ListAll 返回区服列表；serverIDs 为空时返回全部。
func (r *GameServerRepo) ListAll(ctx context.Context, serverIDs []int) ([]model.GameServer, error) {
	filter := bson.M{}
	if len(serverIDs) > 0 {
		filter["server_id"] = bson.M{"$in": serverIDs}
	}
	return r.findAll(ctx, filter, bson.D{{Key: "server_id", Value: 1}})
}

func (r *GameServerRepo) Update(ctx context.Context, id primitive.ObjectID, set bson.M) error {
	return r.updateByID(ctx, id, set)
}

func (r *GameServerRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	return r.deleteByID(ctx, id)
}
