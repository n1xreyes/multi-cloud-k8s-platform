package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// DeploymentLog represents a log entry for application deployments
type DeploymentLog struct {
	ID              primitive.ObjectID     `bson:"_id,omitempty" json:"id,omitempty"`
	ApplicationName string                 `bson:"applicationName" json:"applicationName"`
	Namespace       string                 `bson:"namespace" json:"namespace"`
	ClusterName     string                 `bson:"clusterName" json:"clusterName"`
	Timestamp       time.Time              `bson:"timestamp" json:"timestamp"`
	Level           string                 `bson:"level" json:"level"`
	Message         string                 `bson:"message" json:"message"`
	Details         map[string]interface{} `bson:"details,omitempty" json:"details,omitempty"`
}
