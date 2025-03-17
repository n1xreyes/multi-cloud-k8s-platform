package models

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
