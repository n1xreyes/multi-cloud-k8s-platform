#!/bin/bash

# migrate.sh - Database migration script

set -e

# Configuration
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_NAME=${DB_NAME:-k8s_platform}
MIGRATIONS_PATH=${MIGRATIONS_PATH:-./migrations/postgres}

# Build connection string
DB_URL="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"

# Check if golang-migrate is installed
if ! command -v migrate &> /dev/null; then
    echo "golang-migrate is not installed, installing..."
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
fi

# Command selection
case "$1" in
    up)
        echo "Applying migrations..."
        migrate -path $MIGRATIONS_PATH -database $DB_URL up
        ;;
    down)
        echo "Rolling back all migrations..."
        migrate -path $MIGRATIONS_PATH -database $DB_URL down
        ;;
    reset)
        echo "Resetting database..."
        migrate -path $MIGRATIONS_PATH -database $DB_URL drop
        migrate -path $MIGRATIONS_PATH -database $DB_URL up
        ;;
    create)
        if [ -z "$2" ]; then
            echo "Please provide a migration name"
            exit 1
        fi
        echo "Creating new migration: $2"
        migrate create -ext sql -dir $MIGRATIONS_PATH -seq $2
        ;;
    *)
        echo "Usage: $0 {up|down|reset|create <name>}"
        exit 1
        ;;
esac

echo "Migration operation completed successfully"
