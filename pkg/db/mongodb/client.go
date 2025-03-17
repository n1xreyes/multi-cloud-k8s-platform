package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

// Client wraps the MongoDB client and database
type Client struct {
	client   *mongo.Client
	Database *mongo.Database
}

// Config contains MongoDB connection parameters
type Config struct {
	URI      string
	Database string
}

// NewClient initializes a MongoDB client with connection pooling and monitoring
func NewClient(ctx context.Context, config Config) (*Client, error) {
	opts := options.Client().
		ApplyURI(config.URI).
		SetMaxPoolSize(100).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(30 * time.Second).
		SetServerMonitor(&event.ServerMonitor{
			ServerHeartbeatSucceeded: func(event *event.ServerHeartbeatSucceededEvent) {
				// Add monitoring logic here
			},
			ServerHeartbeatFailed: func(event *event.ServerHeartbeatFailedEvent) {
				// Add monitoring logic here
			},
		})

	// Add retry logic for connection
	var client *mongo.Client
	var err error
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		client, err = mongo.Connect(ctx, opts)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB after %d retries: %w", maxRetries, err)
	}

	// Add timeout for connection check
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("error pinging MongoDB: %w", err)
	}

	return &Client{
		client:   client,
		Database: client.Database(config.Database),
	}, nil
}

// Close terminates the database connection
func (c *Client) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}
