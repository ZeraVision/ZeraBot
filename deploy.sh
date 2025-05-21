#!/bin/bash
set -e

# Load environment variables
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

# Build the Docker image
echo "🚀 Building ZeraBot Docker image..."
docker-compose build

# Stop and remove existing containers if they exist
echo "🛑 Stopping and removing any existing containers..."
docker-compose down --remove-orphans || true

# Start the services in detached mode
echo "🚀 Starting ZeraBot services..."
docker-compose up -d

echo "✅ Deployment complete!"
echo "📝 Check logs with: docker-compose logs -f"
