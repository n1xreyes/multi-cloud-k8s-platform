FROM golang:1.24-alpine AS builder
LABEL authors="Angelo Reyes"

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o auth-server ./cmd/auth/main.go

# Use a small alpine image for the final image
FROM alpine:3.18

RUN apk update && apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/auth-server .

# Copy migrations
COPY --from=builder /app/migrations /app/migrations

# Run the binary
ENTRYPOINT ["/app/auth-server"]