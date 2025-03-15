package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
	cluster.ID = primitive.NewObjectID()
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
	return r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
}

// DeleteCluster removes a cluster by ID
func (r *ClusterRepository) DeleteCluster(ctx context.Context, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	return r.collection.DeleteOne(ctx, bson.M{"_id": id})
}
