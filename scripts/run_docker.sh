#!/bin/bash
# Development script to run with Docker Compose

set -e

echo "Starting services..."
docker-compose up -d

echo "Services started:"
echo "  - API Gateway: http://localhost:8080"
echo "  - PostgreSQL: localhost:5432"

echo ""
echo "View logs:"
echo "  docker-compose logs -f gateway"

echo ""
echo "Stop services:"
echo "  docker-compose down"