package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

// DeploymentLogRepository handles deployment log-related DB operations
type DeploymentLogRepository struct {
	collection *mongo.Collection
}

// NewDeploymentLogRepository initializes a DeploymentLogRepository
func NewDeploymentLogRepository(db *mongo.Database) *DeploymentLogRepository {
	return &DeploymentLogRepository{
		collection: db.Collection("deploymentLogs"),
	}
}

// InsertDeploymentLog adds a new log
func (r *DeploymentLogRepository) InsertDeploymentLog(ctx context.Context, log *DeploymentLog) (*mongo.InsertOneResult, error) {
	// Validation: Ensure required fields are set
	if log.ApplicationName == "" ||
		log.Namespace == "" ||
		log.ClusterName == "" ||
		log.Level == "" ||
		log.Message == "" {
		return nil, fmt.Errorf("missing required fields: ApplicationName, Namespace, ClusterName, Level")
	}

	log.ID = primitive.NewObjectID()
	log.Timestamp = time.Now()

	// Ensure Details is initiated if missing
	if log.Details == nil {
		log.Details = make(map[string]interface{})
	}

	// Insert the document
	return r.collection.InsertOne(ctx, log)
}

// GetDeploymentLog retrieves a deployment log by ID
//func (r *DeploymentLogRepository) GetDeploymentLog(ctx context.Context, id string) (*DeploymentLog, error) {
//	var deploymentLog DeploymentLog
//
//}
