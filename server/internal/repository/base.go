package repository

import (
	"context"
	"errors"
	"regexp"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrNotFound  = errors.New("document not found")
	ErrDuplicate = errors.New("duplicate key")
)

// base 封装单个集合的通用 CRUD，具体 repo 通过嵌入复用。
type base[T any] struct {
	coll *mongo.Collection
}

func translate(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, mongo.ErrNoDocuments):
		return ErrNotFound
	case mongo.IsDuplicateKeyError(err):
		return ErrDuplicate
	default:
		return err
	}
}

func (b base[T]) insert(ctx context.Context, doc *T) error {
	_, err := b.coll.InsertOne(ctx, doc)
	return translate(err)
}

func (b base[T]) findOne(ctx context.Context, filter any) (*T, error) {
	var doc T
	if err := b.coll.FindOne(ctx, filter).Decode(&doc); err != nil {
		return nil, translate(err)
	}
	return &doc, nil
}

func (b base[T]) findByID(ctx context.Context, id primitive.ObjectID) (*T, error) {
	return b.findOne(ctx, bson.M{"_id": id})
}

func (b base[T]) findAll(ctx context.Context, filter any, sort bson.D) ([]T, error) {
	cur, err := b.coll.Find(ctx, filter, options.Find().SetSort(sort))
	if err != nil {
		return nil, err
	}
	list := make([]T, 0)
	if err := cur.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (b base[T]) findPage(ctx context.Context, filter any, sort bson.D, page, size int64) ([]T, int64, error) {
	total, err := b.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	opts := options.Find().SetSort(sort).SetSkip((page - 1) * size).SetLimit(size)
	cur, err := b.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	list := make([]T, 0, size)
	if err := cur.All(ctx, &list); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (b base[T]) updateByID(ctx context.Context, id primitive.ObjectID, set bson.M) error {
	res, err := b.coll.UpdateByID(ctx, id, bson.M{"$set": set})
	if err != nil {
		return translate(err)
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (b base[T]) deleteByID(ctx context.Context, id primitive.ObjectID) error {
	res, err := b.coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (b base[T]) createIndexes(ctx context.Context, models ...mongo.IndexModel) error {
	_, err := b.coll.Indexes().CreateMany(ctx, models)
	return err
}

// fuzzy 构造大小写不敏感的模糊匹配，输入会被转义以避免正则注入。
func fuzzy(keyword string) primitive.Regex {
	return primitive.Regex{Pattern: regexp.QuoteMeta(keyword), Options: "i"}
}
