package repository

import (
	"context"
	"fmt"
	"github.com/n1xreyes/multi-cloud-k8s-platform/pkg/db/mongodb"
	"github.com/n1xreyes/multi-cloud-k8s-platform/pkg/db/mongodb/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

// MetricsRepository handles metrics-related operations
type MetricsRepository struct {
	collection *mongo.Collection
}

// NewMetricsRepository initializes a MetricsRepository
func NewMetricsRepository(client *mongodb.Client) *MetricsRepository {
	return &MetricsRepository{
		collection: client.Database.Collection("metrics"),
	}
}

// InsertMetric inserts a new metric
func (r *MetricsRepository) InsertMetric(ctx context.Context, metric *models.Metric) (*mongo.InsertOneResult, error) {
	// Validate required fields
	if metric.ResourceType == "" || metric.ResourceName == "" || metric.Namespace == "" {
		return nil, fmt.Errorf("missing required fields: ResourceType, ResourceName, and Namespace are required")
	}

	// Ensure ID and timestamp are set
	metric.ID = primitive.NewObjectID()
	metric.Timestamp = time.Now()

	// Insert new metric entry
	return r.collection.InsertOne(ctx, metric)
}

// ListMetrics retrieves metrics over time
func (r *MetricsRepository) ListMetrics(ctx context.Context, filter bson.M) ([]*models.Metric, error) {
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var metrics []*models.Metric
	if err := cursor.All(ctx, &metrics); err != nil {
		return nil, err
	}
	return metrics, nil
}
