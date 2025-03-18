package models

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// use a single instance of Validate, it caches struct info
var clusterValidate *validator.Validate

// Common validation errors
var (
	ErrEmptyClusterName = errors.New("cluster name cannot be empty")
	ErrEmptyProvider    = errors.New("provider cannot be empty")
	ErrInvalidProvider  = errors.New("invalid provider, must be one of: aws, gcp, azure")
	ErrEmptyRegion      = errors.New("region cannot be empty")
)

// Cluster represents a Kubernetes cluster in the system
type Cluster struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string             `bson:"name" json:"name" validate:"required"`
	Namespace string             `bson:"namespace" json:"namespace" validate:"required"`
	UserID    string             `bson:"userId,omitempty" json:"userId,omitempty"`
	Spec      ClusterSpec        `bson:"spec" json:"spec" validate:"required"`
	Status    ClusterStatus      `bson:"status" json:"status"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type NodeGroup struct {
	Name         string            `bson:"name" json:"name" validate:"required"`
	InstanceType string            `bson:"instanceType" json:"instanceType" validate:"required"`
	MinSize      int               `bson:"minSize" json:"minSize" validate:"min=1"`
	MaxSize      int               `bson:"maxSize" json:"maxSize" validate:"required"`
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
	Provider       string          `bson:"provider" json:"provider" validate:"oneof=aws gcp azure"`
	Region         string          `bson:"region" json:"region" validate:"required"`
	Version        string          `bson:"version,omitempty" json:"version,omitempty"`
	NodeGroups     []NodeGroup     `bson:"nodeGroups,omitempty" json:"nodeGroups,omitempty"`
	Networking     *Networking     `bson:"networking,omitempty" json:"networking,omitempty"`
	Authentication *Authentication `bson:"authentication,omitempty" json:"authentication,omitempty"`
}

type ClusterCondition struct {
	Type               string    `bson:"type" json:"type" validate:"required"`
	Status             string    `bson:"status" json:"status" validate:"oneof=True False Unknown"`
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

func (c *Cluster) Validate() error {
	clusterValidate = validator.New(validator.WithRequiredStructEnabled())

	err := clusterValidate.Struct(c)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		for _, fieldErr := range validationErrors {
			switch fieldErr.StructField() {
			case "Name":
				return ErrEmptyClusterName
			case "Provider":
				if fieldErr.Tag() == "required" {
					return ErrEmptyProvider
				}
				if fieldErr.Tag() == "oneof" {
					return ErrInvalidProvider
				}
			case "Region":
				return ErrEmptyRegion
			}
		}
	}

	return nil
}
