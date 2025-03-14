// MongoDB is schema-less, but we'll define the expected document structures

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