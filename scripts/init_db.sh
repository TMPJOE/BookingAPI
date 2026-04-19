#!/bin/bash
# Database initialization script

set -e

echo "Waiting for database to be ready..."
sleep 5

echo "Running migrations..."
# Add migration commands here
# Example: migrate -path ./migrations -database "$DB_URL" up

echo "Database initialization complete"