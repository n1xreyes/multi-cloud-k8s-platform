package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

// Client wraps the MongoDB client and database
type Client struct {
	client   *mongo.Client
	database *mongo.Database
}

// Config contains MongoDB connection parameters
type Config struct {
	URI      string
	Database string
}

// NewClient initializes a MongoDB connection
func NewClient(ctx context.Context, config Config) (*Client, error) {
	opts := options.Client().ApplyURI(config.URI)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("error connecting to MongoDB: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("error pinging MongoDB: %w", err)
	}

	return &Client{
		client:   client,
		database: client.Database(config.Database),
	}, nil
}

// Close terminates the database connection
func (c *Client) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}
