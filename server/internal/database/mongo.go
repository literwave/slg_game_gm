package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"slg_game_gm/server/internal/config"
)

// Connect 建立 MongoDB 连接并做一次 Ping 校验。
func Connect(ctx context.Context, cfg config.MongoConfig) (*mongo.Client, *mongo.Database, error) {
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	opts := options.Client().
		ApplyURI(cfg.URI).
		SetConnectTimeout(timeout).
		SetServerSelectionTimeout(timeout)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("connect mongo: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, nil, fmt.Errorf("ping mongo: %w", err)
	}
	return client, client.Database(cfg.Database), nil
}
