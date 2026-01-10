#!/bin/sh
set -e

# V Panel Docker Entrypoint Script

echo "Starting V Panel..."

# Create config from example if not exists
if [ ! -f /app/configs/config.yaml ]; then
    echo "Creating default configuration..."
    cp /app/configs/config.yaml.example /app/configs/config.yaml
fi

# Ensure data directory exists and has correct permissions
mkdir -p /app/data /app/logs

# Initialize database if needed
if [ ! -f /app/data/v.db ]; then
    echo "Initializing database..."
    touch /app/data/v.db
fi

# Print startup information
echo "Configuration:"
echo "  Server Host: ${V_SERVER_HOST:-0.0.0.0}"
echo "  Server Port: ${V_SERVER_PORT:-8080}"
echo "  Log Level: ${V_LOG_LEVEL:-info}"
echo "  Database: ${V_DB_PATH:-/app/data/v.db}"

# Execute the main command
exec "$@"
