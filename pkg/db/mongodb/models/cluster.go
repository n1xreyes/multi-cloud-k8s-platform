package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

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
