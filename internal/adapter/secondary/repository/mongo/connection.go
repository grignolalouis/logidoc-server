package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/logidoc/logidoc-server/internal/config"
)

type Connection struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewConnection(cfg config.MongoConfig) (*Connection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("mongo ping (is MongoDB running at %s?): %w", cfg.URI, err)
	}

	return &Connection{
		client: client,
		db:     client.Database(cfg.Database),
	}, nil
}

func (c *Connection) Database() *mongo.Database { return c.db }

func (c *Connection) Check(ctx context.Context) error {
	return c.client.Ping(ctx, nil)
}

func (c *Connection) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}
