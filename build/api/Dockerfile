
FROM golang:1.21-alpine as builder
LABEL authors="Angelo Reyes"

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o api-server ./cmd/api-server/main.go

# Use a small alpine image for the final image
FROM alpine:3.18

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/api-server .

# Copy migrations
COPY --from=builder /app/migrations /app/migrations

# Run the binary
ENTRYPOINT ["/app/api-server"]