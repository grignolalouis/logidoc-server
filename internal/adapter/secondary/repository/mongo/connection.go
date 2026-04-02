package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/logidoc/logidoc-server/internal/config"
)

// Connection holds the MongoDB client and database.
type Connection struct {
	client *mongo.Client
	db     *mongo.Database
}

// NewConnection creates a new MongoDB connection.
func NewConnection(cfg config.MongoConfig) (*Connection, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}

	if err := client.Ping(context.Background(), nil); err != nil {
		return nil, fmt.Errorf("mongo ping: %w", err)
	}

	return &Connection{
		client: client,
		db:     client.Database(cfg.Database),
	}, nil
}

// Database returns the MongoDB database.
func (c *Connection) Database() *mongo.Database {
	return c.db
}

// Close disconnects the MongoDB client.
func (c *Connection) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}
