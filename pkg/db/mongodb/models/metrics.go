package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// Metric represents a time-series metrics data point
type Metric struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ResourceType string             `bson:"resourceType" json:"resourceType"`
	ResourceName string             `bson:"resourceName" json:"resourceName"`
	Namespace    string             `bson:"namespace" json:"namespace"`
	ClusterName  string             `bson:"clusterName,omitempty" json:"clusterName,omitempty"`
	Timestamp    time.Time          `bson:"timestamp" json:"timestamp"`
	Metrics      MetricsData        `bson:"metrics" json:"metrics"`
}

type CPUMetrics struct {
	Usage    float64 `bson:"usage,omitempty" json:"usage,omitempty"`
	Limit    float64 `bson:"limit,omitempty" json:"limit,omitempty"`
	Requests float64 `bson:"requests,omitempty" json:"requests,omitempty"`
}

type MemoryMetrics struct {
	Usage    float64 `bson:"usage,omitempty" json:"usage,omitempty"`
	Limit    float64 `bson:"limit,omitempty" json:"limit,omitempty"`
	Requests float64 `bson:"requests,omitempty" json:"requests,omitempty"`
}

type NetworkMetrics struct {
	RxBytes  int64 `bson:"rxBytes,omitempty" json:"rxBytes,omitempty"`
	TxBytes  int64 `bson:"txBytes,omitempty" json:"txBytes,omitempty"`
	RxErrors int   `bson:"rxErrors,omitempty" json:"rxErrors,omitempty"`
	TxErrors int   `bson:"txErrors,omitempty" json:"txErrors,omitempty"`
}

type ReplicaMetrics struct {
	Desired   int `bson:"desired,omitempty" json:"desired,omitempty"`
	Ready     int `bson:"ready,omitempty" json:"ready,omitempty"`
	Available int `bson:"available,omitempty" json:"available,omitempty"`
}

// MetricsData contains the actual metrics measurements
type MetricsData struct {
	CPU      *CPUMetrics     `bson:"cpu,omitempty" json:"cpu,omitempty"`
	Memory   *MemoryMetrics  `bson:"memory,omitempty" json:"memory,omitempty"`
	Network  *NetworkMetrics `bson:"network,omitempty" json:"network,omitempty"`
	Replicas *ReplicaMetrics `bson:"replicas,omitempty" json:"replicas,omitempty"`
}
