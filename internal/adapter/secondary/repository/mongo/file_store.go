package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type fileDoc struct {
	ID   string `bson:"_id"`
	Data []byte `bson:"data"`
}

// FileStore stores raw file bytes in MongoDB.
type FileStore struct {
	col *mongo.Collection
}

// NewFileStore creates a new FileStore.
func NewFileStore(conn *Connection) *FileStore {
	return &FileStore{col: conn.Database().Collection("files")}
}

func (s *FileStore) Save(ctx context.Context, id string, data []byte) error {
	_, err := s.col.ReplaceOne(ctx,
		bson.M{"_id": id},
		fileDoc{ID: id, Data: data},
		options.Replace().SetUpsert(true),
	)
	return err
}

func (s *FileStore) Load(ctx context.Context, id string) ([]byte, error) {
	var doc fileDoc
	if err := s.col.FindOne(ctx, bson.M{"_id": id}).Decode(&doc); err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}
	return doc.Data, nil
}

func (s *FileStore) Delete(ctx context.Context, id string) error {
	_, err := s.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
