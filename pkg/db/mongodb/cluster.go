package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

// ClusterRepository handles cluster-related DB operations
type ClusterRepository struct {
	collection *mongo.Collection
}

// NewClusterRepository initializes a ClusterRepository
func NewClusterRepository(client *Client) *ClusterRepository {
	return &ClusterRepository{
		collection: client.database.Collection("clusters"),
	}
}

// InsertCluster adds a new cluster
func (r *ClusterRepository) InsertCluster(ctx context.Context, cluster *Cluster) (*mongo.InsertOneResult, error) {
	// Validate required fields
	if cluster.Name == "" || cluster.Namespace == "" {
		return nil, fmt.Errorf("missing required fields: Name and Namespace are required")
	}
	if cluster.Spec.Provider == "" || cluster.Spec.Region == "" {
		return nil, fmt.Errorf("missing required fields: Spec.Provider and Spec.Region are required")
	}

	// Ensure ID and timestamps are set
	cluster.ID = primitive.NewObjectID()
	now := time.Now()
	cluster.CreatedAt = now
	cluster.UpdatedAt = now

	// Initialize nested fields to prevent nil values
	if cluster.Spec.NodeGroups == nil {
		cluster.Spec.NodeGroups = []NodeGroup{}
	}
	if cluster.Spec.Networking == nil {
		cluster.Spec.Networking = &Networking{}
	}
	if cluster.Spec.Authentication == nil {
		cluster.Spec.Authentication = &Authentication{}
	}
	if cluster.Status.Conditions == nil {
		cluster.Status.Conditions = []ClusterCondition{}
	}

	// Insert into MongoDB
	return r.collection.InsertOne(ctx, cluster)
}

// GetCluster retrieves a cluster by ID
func (r *ClusterRepository) GetCluster(ctx context.Context, id primitive.ObjectID) (*Cluster, error) {
	var cluster Cluster
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&cluster)
	return &cluster, err
}

// UpdateCluster updates a cluster
func (r *ClusterRepository) UpdateCluster(ctx context.Context, id primitive.ObjectID, update bson.M) (*mongo.UpdateResult, error) {
	// Fetch existing cluster
	var existingCluster Cluster
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&existingCluster)
	if err != nil {
		return nil, fmt.Errorf("cluster not found")
	}

	// Validate required fields
	if name, ok := update["name"].(string); ok && name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	if namespace, ok := update["namespace"].(string); ok && namespace == "" {
		return nil, fmt.Errorf("namespace cannot be empty")
	}
	if spec, ok := update["spec"].(bson.M); ok {
		if provider, provOk := spec["provider"].(string); provOk && provider == "" {
			return nil, fmt.Errorf("spec.provider cannot be empty")
		}
		if region, regOk := spec["region"].(string); regOk && region == "" {
			return nil, fmt.Errorf("spec.region cannot be empty")
		}
	}

	// Ensure update uses $set
	if _, exists := update["$set"]; !exists {
		update["$set"] = bson.M{}
	}
	setFields := update["$set"].(bson.M)

	// Preserve existing values if not provided in the update
	if _, ok := setFields["spec"]; !ok {
		setFields["spec"] = bson.M{}
	}
	specFields := setFields["spec"].(bson.M)

	if _, ok := specFields["nodeGroups"]; !ok {
		specFields["nodeGroups"] = existingCluster.Spec.NodeGroups
	}
	if _, ok := specFields["networking"]; !ok {
		specFields["networking"] = existingCluster.Spec.Networking
	}
	if _, ok := specFields["authentication"]; !ok {
		specFields["authentication"] = existingCluster.Spec.Authentication
	}
	setFields["spec"] = specFields

	if _, ok := setFields["status"]; !ok {
		setFields["status"] = bson.M{}
	}
	statusFields := setFields["status"].(bson.M)

	if _, ok := statusFields["conditions"]; !ok {
		statusFields["conditions"] = existingCluster.Status.Conditions
	}
	setFields["status"] = statusFields

	// Always update the timestamp
	setFields["updatedAt"] = time.Now()

	// Perform update
	return r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
}

// DeleteCluster removes a cluster by ID
func (r *ClusterRepository) DeleteCluster(ctx context.Context, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	return r.collection.DeleteOne(ctx, bson.M{"_id": id})
}
