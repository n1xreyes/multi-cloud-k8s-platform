package models

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
	"time"
)

// Use a single instance of Validate, it caches struct info
var appValidate *validator.Validate

// Common validation errors
var (
	ErrEmptyAppName  = errors.New("application name cannot be empty")
	ErrEmptyImage    = errors.New("image cannot be empty")
	ErrInvalidStatus = errors.New("invalid status, must be one of: True, False, Unknown")
)

// Allowed statuses for case-insensitive validation
var validStatuses = map[string]bool{
	"true":    true,
	"false":   true,
	"unknown": true,
}

// Application represents an application deployment in the system
type Application struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string             `bson:"name" json:"name" validate:"required"`
	Namespace string             `bson:"namespace" json:"namespace" validate:"required"`
	UserID    string             `bson:"userId,omitempty" json:"userId,omitempty"`
	Spec      ApplicationSpec    `bson:"spec" json:"spec" validate:"required"`
	Status    ApplicationStatus  `bson:"status" json:"status" validate:"required"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// ApplicationSpec contains the desired state of the application
type ApplicationSpec struct {
	Image              string     `bson:"image" json:"image" validate:"required"`
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

type ApplicationCondition struct {
	Type               string    `bson:"type" json:"type" validate:"required"`
	Status             string    `bson:"status" json:"status" validate:"oneof=True False Unknown"` // Enum: ["True", "False", "Unknown"]
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

// Custom validator function for Status field
func validateStatus(fl validator.FieldLevel) bool {
	status := strings.ToLower(fl.Field().String())
	return validStatuses[status]
}

// Initialize validator instance and register custom validations
func init() {
	appValidate = validator.New(validator.WithRequiredStructEnabled())

	// Register custom validation for status field
	_ = appValidate.RegisterValidation("appStatus", validateStatus)
}

func (a *Application) Validate() error {
	err := appValidate.Struct(a)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		for _, fieldErr := range validationErrors {
			switch fieldErr.StructField() {
			case "Name":
				return ErrEmptyAppName
			case "Image":
				return ErrEmptyImage
			case "Status":
				if fieldErr.Tag() == "oneof" {
					return ErrInvalidStatus
				}
			}
		}
	}

	return nil
}
