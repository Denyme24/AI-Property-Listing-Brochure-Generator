package services

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBService struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func NewMongoDBService(uri, database string) (*MongoDBService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return &MongoDBService{
		Client:   client,
		Database: client.Database(database),
	}, nil
}

func (s *MongoDBService) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.Client.Disconnect(ctx)
}

func (s *MongoDBService) GetCollection(name string) *mongo.Collection {
	return s.Database.Collection(name)
}

