FROM golang:1.24-alpine AS builder
LABEL authors="Angelo Reyes"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api-gateway ./cmd/api-gateway/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /bin/api-gateway .

EXPOSE 8080

CMD ["./api-gateway"]
