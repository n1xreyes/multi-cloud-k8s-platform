package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

// ApplicationRepository handles application-related DB operations
type ApplicationRepository struct {
	collection *mongo.Collection
}

// NewApplicationRepository initializes an ApplicationRepository
func NewApplicationRepository(client *Client) *ApplicationRepository {
	return &ApplicationRepository{
		collection: client.database.Collection("applications"),
	}
}

// InsertApplication adds a new application
func (r *ApplicationRepository) InsertApplication(ctx context.Context, app *Application) (*mongo.InsertOneResult, error) {
	// Validation: Ensure required fields are set
	if app.Name == "" || app.Namespace == "" {
		return nil, fmt.Errorf("missing required fields: Name and Namespace are required")
	}

	if app.Spec.Image == "" {
		return nil, fmt.Errorf("missing required field: Spec.Image is required")
	}

	// Ensure ID and timestamps are set
	app.ID = primitive.NewObjectID()
	now := time.Now()
	app.CreatedAt = now
	app.UpdatedAt = now

	// Ensure Spec and Status are initialized if missing
	if app.Spec.TargetClusters == nil {
		app.Spec.TargetClusters = []string{}
	}
	if app.Status.Conditions == nil {
		app.Status.Conditions = []ApplicationCondition{}
	}
	if app.Status.Deployments == nil {
		app.Status.Deployments = []DeploymentStatus{}
	}

	// Insert the document
	return r.collection.InsertOne(ctx, app)
}

// GetApplication retrieves an application by ID
func (r *ApplicationRepository) GetApplication(ctx context.Context, id primitive.ObjectID) (*Application, error) {
	var app Application
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&app)
	return &app, err
}

// UpdateApplication updates an application
func (r *ApplicationRepository) UpdateApplication(ctx context.Context, id primitive.ObjectID, update bson.M) (*mongo.UpdateResult, error) {
	// Fetch the current application to preserve existing values
	var existingApp Application
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&existingApp)
	if err != nil {
		return nil, fmt.Errorf("application not found")
	}

	// Validate required fields
	if name, ok := update["name"].(string); ok && name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	if namespace, ok := update["namespace"].(string); ok && namespace == "" {
		return nil, fmt.Errorf("namespace cannot be empty")
	}
	if spec, ok := update["spec"].(bson.M); ok {
		if image, imgOk := spec["image"].(string); imgOk && image == "" {
			return nil, fmt.Errorf("spec.image cannot be empty")
		}
	}

	// Ensure update uses $set
	if _, exists := update["$set"]; !exists {
		update["$set"] = bson.M{}
	}
	setFields := update["$set"].(bson.M)

	// Set default values if missing
	if _, ok := setFields["spec"]; !ok {
		setFields["spec"] = bson.M{}
	}
	specFields := setFields["spec"].(bson.M)

	if _, ok := specFields["targetClusters"]; !ok {
		specFields["targetClusters"] = existingApp.Spec.TargetClusters
	}
	setFields["spec"] = specFields // Reassign to ensure it stays in $set

	if _, ok := setFields["status"]; !ok {
		setFields["status"] = bson.M{}
	}
	statusFields := setFields["status"].(bson.M)

	if _, ok := statusFields["conditions"]; !ok {
		statusFields["conditions"] = existingApp.Status.Conditions
	}
	if _, ok := statusFields["deployments"]; !ok {
		statusFields["deployments"] = existingApp.Status.Deployments
	}
	setFields["status"] = statusFields

	// Always update the timestamp
	setFields["updatedAt"] = time.Now()

	// Perform update
	return r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
}

// DeleteApplication removes an application by ID
func (r *ApplicationRepository) DeleteApplication(ctx context.Context, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	return r.collection.DeleteOne(ctx, bson.M{"_id": id})
}

// ListApplications fetches all applications
func (r *ApplicationRepository) ListApplications(ctx context.Context, filter bson.M) ([]Application, error) {
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var apps []Application
	if err = cursor.All(ctx, &apps); err != nil {
		return nil, err
	}
	return apps, nil
}
