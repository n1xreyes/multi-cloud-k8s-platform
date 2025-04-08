FROM golang:1.24-alpine AS builder
LABEL authors="Angelo Reyes"

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Build config-server binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/config-server ./cmd/config-server/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKIR /root/

# Copy the binary from the builder stage
COPY --from=builder /bin/config-server .

EXPOSE 8082

CMD ["./config-server"]