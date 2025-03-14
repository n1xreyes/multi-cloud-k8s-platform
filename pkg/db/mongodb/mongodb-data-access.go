package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

// Config contains the MongoDB connection parameters
type Config struct {
	URI      string
	Database string
}

// Client represents a MongoDB client
type Client struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewClient creates a new MongoDB client
func NewClient(ctx context.Context, config Config) (*Client, error) {
	opts := options.Client().ApplyURI(config.URI)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("error connecting to MongoDB: %w", err)
	}

	// Verify the connection
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("error pinging MongoDB: %w", err)
	}

	database := client.Database(config.Database)

	return &Client{
		client:   client,
		database: database,
	}, nil
}

// Close closes the MongoDB connection
func (c *Client) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

// Application represents an application deployment in the system
type Application struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string             `bson:"name" json:"name"`
	Namespace string             `bson:"namespace" json:"namespace"`
	UserID    string             `bson:"userId,omitempty" json:"userId,omitempty"`
	Spec      ApplicationSpec    `bson:"spec" json:"spec"`
	Status    ApplicationStatus  `bson:"status" json:"status"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// ApplicationSpec contains the desired state of the application
type ApplicationSpec struct {
	Image              string     `bson:"image" json:"image"`
	Replicas           *int32     `bson:"replicas,omitempty" json:"replicas,omitempty"`
	Resources          *Resources `bson:"resources,omitempty" json:"resources,omitempty"`
	Ports              []Port     `bson:"ports,omitempty" json:"ports,omitempty"`
	Env                []EnvVar   `bson:"env,omitempty" json:"env,omitempty"`
	TargetClusters     []string   `bson:"targetClusters,omitempty" json:"targetClusters,omitempty"`
	DeploymentStrategy string     `bson:"deploymentStrategy,omitempty" json:"deploymentStrategy,omitempty"`
}

// Resources contains the cpu and memory requests of the application
type Resources struct {
	CPU    string `bson:"cpu,omitempty" json:"cpu,omitempty"`
	Memory string `bson:"memory,omitempty" json:"memory,omitempty"`
}

// Port represents a container port
type Port struct {
	Name          string `bson:"name,omitempty" json:"name,omitempty"`
	ContainerPort int32  `bson:"containerPort,omitempty" json:"containerPort,omitempty"`
	Protocol      string `bson:"protocol,omitempty" json:"protocol,omitempty"`
}

// EnvVar contains the environment variables, config maps and secret keys for the application
type EnvVar struct {
	Name      string     `bson:"name" json:"name"`
	Value     string     `bson:"value,omitempty" json:"value,omitempty"`
	ValueFrom *ValueFrom `bson:"valueFrom,omitempty" json:"valueFrom,omitempty"`
}

// ConfigMapKeyRef represents the config map for the application
type ConfigMapKeyRef struct {
	Name string `bson:"name,omitempty" json:"name,omitempty"`
	Key  string `bson:"key,omitempty" json:"key,omitempty"`
}

// SecretKeyRef represents the secret key references for the application
type SecretKeyRef struct {
	Name string `bson:"name,omitempty" json:"name,omitempty"`
	Key  string `bson:"key,omitempty" json:"key,omitempty"`
}

// ValueFrom contains key value maps for the config variables and secret keys of the application
type ValueFrom struct {
	ConfigMapKeyRef *ConfigMapKeyRef `bson:"configMapKeyRef,omitempty" json:"configMapKeyRef,omitempty"`
	SecretKeyRef    *SecretKeyRef    `bson:"secretKeyRef,omitempty" json:"secretKeyRef,omitempty"`
}

type ApplicationCondition struct {
	Type               string    `bson:"type" json:"type"`
	Status             string    `bson:"status" json:"status"` // Enum: ["True", "False", "Unknown"]
	LastTransitionTime time.Time `bson:"lastTransitionTime,omitempty" json:"lastTransitionTime,omitempty"`
	Reason             string    `bson:"reason,omitempty" json:"reason,omitempty"`
	Message            string    `bson:"message,omitempty" json:"message,omitempty"`
}

type DeploymentStatus struct {
	Cluster     string    `bson:"cluster,omitempty" json:"cluster,omitempty"`
	Status      string    `bson:"status,omitempty" json:"status,omitempty"`
	Ready       bool      `bson:"ready,omitempty" json:"ready,omitempty"`
	LastUpdated time.Time `bson:"lastUpdated,omitempty" json:"lastUpdated,omitempty"`
}

// ApplicationStatus contains the current state of the application
type ApplicationStatus struct {
	ObservedGeneration int32                  `bson:"observedGeneration,omitempty" json:"observedGeneration,omitempty"`
	Conditions         []ApplicationCondition `bson:"conditions,omitempty" json:"conditions,omitempty"`
	Deployments        []DeploymentStatus     `bson:"deployments,omitempty" json:"deployments,omitempty"`
}

// Cluster represents a Kubernetes cluster in the system
type Cluster struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string             `bson:"name" json:"name"`
	Namespace string             `bson:"namespace" json:"namespace"`
	UserID    string             `bson:"userId,omitempty" json:"userId,omitempty"`
	Spec      ClusterSpec        `bson:"spec" json:"spec"`
	Status    ClusterStatus      `bson:"status" json:"status"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type NodeGroup struct {
	Name         string            `bson:"name" json:"name"`
	InstanceType string            `bson:"instanceType" json:"instanceType"`
	MinSize      int               `bson:"minSize" json:"minSize"`
	MaxSize      int               `bson:"maxSize" json:"maxSize"`
	Labels       map[string]string `bson:"labels,omitempty" json:"labels,omitempty"`
}

type Networking struct {
	VpcCidr         string `bson:"vpcCidr,omitempty" json:"vpcCidr,omitempty"`
	SubnetCidr      string `bson:"subnetCidr,omitempty" json:"subnetCidr,omitempty"`
	ServiceIpv4Cidr string `bson:"serviceIpv4Cidr,omitempty" json:"serviceIpv4Cidr,omitempty"`
}

type Authentication struct {
	RoleArn string `bson:"roleArn,omitempty" json:"roleArn,omitempty"`
	UserArn string `bson:"userArn,omitempty" json:"userArn,omitempty"`
}

// ClusterSpec contains the desired state of the cluster
type ClusterSpec struct {
	Provider       string          `bson:"provider" json:"provider"`
	Region         string          `bson:"region" json:"region"`
	Version        string          `bson:"version,omitempty" json:"version,omitempty"`
	NodeGroups     []NodeGroup     `bson:"nodeGroups,omitempty" json:"nodeGroups,omitempty"`
	Networking     *Networking     `bson:"networking,omitempty" json:"networking,omitempty"`
	Authentication *Authentication `bson:"authentication,omitempty" json:"authentication,omitempty"`
}

type ClusterCondition struct {
	Type               string    `bson:"type" json:"type"`
	Status             string    `bson:"status" json:"status"`
	LastTransitionTime time.Time `bson:"lastTransitionTime,omitempty" json:"lastTransitionTime,omitempty"`
	Reason             string    `bson:"reason,omitempty" json:"reason,omitempty"`
	Message            string    `bson:"message,omitempty" json:"message,omitempty"`
}

// ClusterStatus contains the current state of the cluster
type ClusterStatus struct {
	ObservedGeneration int32              `bson:"observedGeneration,omitempty" json:"observedGeneration,omitempty"`
	Conditions         []ClusterCondition `bson:"conditions,omitempty" json:"conditions,omitempty"`
	Kubeconfig         string             `bson:"kubeconfig,omitempty" json:"kubeconfig,omitempty"`
	Endpoint           string             `bson:"endpoint,omitempty" json:"endpoint,omitempty"`
	Status             string             `bson:"status,omitempty" json:"status,omitempty"`
}

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
