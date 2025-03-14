// applications collection - stores application state and deployment information
db.createCollection("applications", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["name", "namespace", "spec", "status"],
      properties: {
        name: {
          bsonType: "string",
          description: "Name of the application"
        },
        namespace: {
          bsonType: "string",
          description: "Kubernetes namespace"
        },
        userId: {
          bsonType: "string",
          description: "ID of the owner user"
        },
        spec: {
          bsonType: "object",
          required: ["image"],
          properties: {
            image: {
              bsonType: "string",
              description: "Container image"
            },
            replicas: {
              bsonType: "int",
              description: "Number of replicas"
            },
            resources: {
              bsonType: "object",
              properties: {
                cpu: {
                  bsonType: "string",
                  description: "CPU request"
                },
                memory: {
                  bsonType: "string",
                  description: "Memory request"
                }
              }
            },
            ports: {
              bsonType: "array",
              items: {
                bsonType: "object",
                required: ["containerPort"],
                properties: {
                  name: {
                    bsonType: "string"
                  },
                  containerPort: {
                    bsonType: "int"
                  },
                  protocol: {
                    bsonType: "string",
                    enum: ["TCP", "UDP"]
                  }
                }
              }
            },
            env: {
              bsonType: "array",
              items: {
                bsonType: "object",
                required: ["name"],
                properties: {
                  name: {
                    bsonType: "string"
                  },
                  value: {
                    bsonType: "string"
                  },
                  valueFrom: {
                    bsonType: "object",
                    properties: {
                      configMapKeyRef: {
                        bsonType: "object",
                        properties: {
                          name: { bsonType: "string" },
                          key: { bsonType: "string" }
                        }
                      },
                      secretKeyRef: {
                        bsonType: "object",
                        properties: {
                          name: { bsonType: "string" },
                          key: { bsonType: "string" }
                        }
                      }
                    }
                  }
                }
              }
            },
            targetClusters: {
              bsonType: "array",
              items: {
                bsonType: "string"
              }
            },
            deploymentStrategy: {
              bsonType: "string",
              enum: ["RollingUpdate", "Recreate", "BlueGreen", "Canary"]
            }
          }
        },
        status: {
          bsonType: "object",
          properties: {
            observedGeneration: {
              bsonType: "int"
            },
            conditions: {
              bsonType: "array",
              items: {
                bsonType: "object",
                required: ["type", "status"],
                properties: {
                  type: {
                    bsonType: "string"
                  },
                  status: {
                    bsonType: "string",
                    enum: ["True", "False", "Unknown"]
                  },
                  lastTransitionTime: {
                    bsonType: "date"
                  },
                  reason: {
                    bsonType: "string"
                  },
                  message: {
                    bsonType: "string"
                  }
                }
              }
            },
            deployments: {
              bsonType: "array",
              items: {
                bsonType: "object",
                properties: {
                  cluster: {
                    bsonType: "string"
                  },
                  status: {
                    bsonType: "string"
                  },
                  ready: {
                    bsonType: "bool"
                  },
                  lastUpdated: {
                    bsonType: "date"
                  }
                }
              }
            }
          }
        },
        createdAt: {
          bsonType: "date",
          description: "Creation timestamp"
        },
        updatedAt: {
          bsonType: "date",
          description: "Last update timestamp"
        }
      }
    }
  }
});

// Create indexes for better query performance
db.applications.createIndex({ "name": 1, "namespace": 1 }, { unique: true });
db.applications.createIndex({ "userId": 1 });
db.applications.createIndex({ "spec.targetClusters": 1 });

// clusters collection - stores the state of managed Kubernetes clusters
db.createCollection("clusters", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["name", "namespace", "spec", "status"],
      properties: {
        name: {
          bsonType: "string",
          description: "Name of the cluster"
        },
        namespace: {
          bsonType: "string",
          description: "Kubernetes namespace"
        },
        userId: {
          bsonType: "string",
          description: "ID of the owner user"
        },
        spec: {
          bsonType: "object",
          required: ["provider", "region"],
          properties: {
            provider: {
              bsonType: "string",
              enum: ["aws", "gcp", "azure"],
              description: "Cloud provider"
            },
            region: {
              bsonType: "string",
              description: "Cloud region"
            },
            version: {
              bsonType: "string",
              description: "Kubernetes version"
            },
            nodeGroups: {
              bsonType: "array",
              items: {
                bsonType: "object",
                required: ["name", "instanceType", "minSize", "maxSize"],
                properties: {
                  name: {
                    bsonType: "string"
                  },
                  instanceType: {
                    bsonType: "string"
                  },
                  minSize: {
                    bsonType: "int"
                  },
                  maxSize: {
                    bsonType: "int"
                  },
                  labels: {
                    bsonType: "object",
                    additionalProperties: {
                      bsonType: "string"
                    }
                  }
                }
              }
            },
            networking: {
              bsonType: "object",
              properties: {
                vpcCidr: {
                  bsonType: "string"
                },
                subnetCidr: {
                  bsonType: "string"
                },
                serviceIpv4Cidr: {
                  bsonType: "string"
                }
              }
            },
            authentication: {
              bsonType: "object",
              properties: {
                roleArn: {
                  bsonType: "string"
                },
                userArn: {
                  bsonType: "string"
                }
              }
            }
          }
        },
        status: {
          bsonType: "object",
          properties: {
            observedGeneration: {
              bsonType: "int"
            },
            conditions: {
              bsonType: "array",
              items: {
                bsonType: "object",
                required: ["type", "status"],
                properties: {
                  type: {
                    bsonType: "string"
                  },
                  status: {
                    bsonType: "string",
                    enum: ["True", "False", "Unknown"]
                  },
                  lastTransitionTime: {
                    bsonType: "date"
                  },
                  reason: {
                    bsonType: "string"
                  },
                  message: {
                    bsonType: "string"
                  }
                }
              }
            },
            kubeconfig: {
              bsonType: "string"
            },
            endpoint: {
              bsonType: "string"
            },
            status: {
              bsonType: "string",
              enum: ["Pending", "Creating", "Running", "Failed", "Deleting", "Deleted"]
            }
          }
        },
        createdAt: {
          bsonType: "date",
          description: "Creation timestamp"
        },
        updatedAt: {
          bsonType: "date",
          description: "Last update timestamp"
        }
      }
    }
  }
});

// Create indexes for better query performance
db.clusters.createIndex({ "name": 1, "namespace": 1 }, { unique: true });
db.clusters.createIndex({ "userId": 1 });
db.clusters.createIndex({ "spec.provider": 1 });
db.clusters.createIndex({ "status.status": 1 });

// deploymentLogs collection - stores detailed deployment logs and events
db.createCollection("deploymentLogs", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["applicationName", "namespace", "clusterName", "timestamp", "level", "message"],
      properties: {
        applicationName: {
          bsonType: "string",
          description: "Name of the application"
        },
        namespace: {
          bsonType: "string",
          description: "Kubernetes namespace"
        },
        clusterName: {
          bsonType: "string",
          description: "Name of the cluster"
        },
        timestamp: {
          bsonType: "date",
          description: "Event timestamp"
        },
        level: {
          bsonType: "string",
          enum: ["INFO", "WARNING", "ERROR", "DEBUG"],
          description: "Log level"
        },
        message: {
          bsonType: "string",
          description: "Log message"
        },
        details: {
          bsonType: "object",
          description: "Additional details about the event"
        }
      }
    }
  }
});

// Create indexes for better query performance and time-based expiration
db.deploymentLogs.createIndex({ "applicationName": 1, "namespace": 1 });
db.deploymentLogs.createIndex({ "clusterName": 1 });
db.deploymentLogs.createIndex({ "timestamp": 1 });
db.deploymentLogs.createIndex({ "level": 1 });
// Create TTL index to automatically expire old logs after 30 days
db.deploymentLogs.createIndex({ "timestamp": 1 }, { expireAfterSeconds: 2592000 });

// metrics collection - stores time-series metrics data for applications and clusters
db.createCollection("metrics", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["resourceType", "resourceName", "namespace", "timestamp", "metrics"],
      properties: {
        resourceType: {
          bsonType: "string",
          enum: ["application", "cluster", "node"],
          description: "Type of resource"
        },
        resourceName: {
          bsonType: "string",
          description: "Name of the resource"
        },
        namespace: {
          bsonType: "string",
          description: "Kubernetes namespace"
        },
        clusterName: {
          bsonType: "string",
          description: "Name of the cluster"
        },
        timestamp: {
          bsonType: "date",
          description: "Metrics timestamp"
        },
        metrics: {
          bsonType: "object",
          description: "Metrics data",
          properties: {
            cpu: {
              bsonType: "object",
              properties: {
                usage: { bsonType: "double" },
                limit: { bsonType: "double" },
                requests: { bsonType: "double" }
              }
            },
            memory: {
              bsonType: "object",
              properties: {
                usage: { bsonType: "double" },
                limit: { bsonType: "double" },
                requests: { bsonType: "double" }
              }
            },
            network: {
              bsonType: "object",
              properties: {
                rxBytes: { bsonType: "long" },
                txBytes: { bsonType: "long" },
                rxErrors: { bsonType: "int" },
                txErrors: { bsonType: "int" }
              }
            },
            replicas: {
              bsonType: "object",
              properties: {
                desired: { bsonType: "int" },
                ready: { bsonType: "int" },
                available: { bsonType: "int" }
              }
            }
          }
        }
      }
    }
  }
});

// Create indexes for better query performance and time-based expiration
db.metrics.createIndex({ "resourceType": 1, "resourceName": 1, "namespace": 1 });
db.metrics.createIndex({ "clusterName": 1 });
db.metrics.createIndex({ "timestamp": 1 });
// Create TTL index to automatically expire old metrics data after 7 days
db.metrics.createIndex({ "timestamp": 1 }, { expireAfterSeconds: 604800 });