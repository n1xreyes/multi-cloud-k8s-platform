#!/bin/bash

set -e

DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_NAME=${DB_NAME:-k8s_platform}
MIGRATIONS_PATH=${MIGRATIONS_PATH:-/app/migrations/postgres}

DB_URL="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"

MIGRATE_CMD="/go/bin/migrate"

echo "Waiting for Postgres database to become ready..."

MAX_WAIT_SECONDS=60
WAIT_INTERVAL=5
SECONDS_WAITED=0

while ! pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME; do
    if [ $SECONDS_WAITED -ge $MAX_WAIT_SECONDS ]; then
        echo "ERROR: Postgres database did not become ready within ${MAX_WAIT_SECONDS} seconds."
        exit 1
    fi
    echo "Postgres not ready yet. Waiting ${WAIT_INTERVAL} more seconds..."
    sleep $WAIT_INTERVAL
    SECONDS_WAITED=$((SECONDS_WAITED + WAIT_INTERVAL))
done

echo "Postgres database is ready."

case "$1" in
    up)
        echo "Applying migrations using $MIGRATE_CMD..."
        "$MIGRATE_CMD" -path "$MIGRATIONS_PATH" -database "$DB_URL" up
        ;;
    down)
        echo "Rolling back all migrations using $MIGRATE_CMD..."
        "$MIGRATE_CMD" -path "$MIGRATIONS_PATH" -database "$DB_URL" down
        ;;
    reset)
        echo "Resetting database using $MIGRATE_CMD..."
        "$MIGRATE_CMD" -path "$MIGRATIONS_PATH" -database "$DB_URL" force -1 || echo "Ignoring drop error (database might not exist)."
        "$MIGRATE_CMD" -path "$MIGRATIONS_PATH" -database "$DB_URL" up
        ;;
    create)
        # This command should still be run locally on the host, not via kubectl exec
        echo "ERROR: 'create' command should be run from the host machine, not inside the container."
        exit 1
        ;;
    *)
        echo "Usage (inside container): $0 {up|down|reset}"
        exit 1
        ;;
esac

echo "Migration operation completed successfully"