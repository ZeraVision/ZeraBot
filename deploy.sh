#!/bin/bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Check if required commands exist
for cmd in docker docker-compose git; do
    if ! command -v "$cmd" > /dev/null; then
        echo -e "${RED}❌ $cmd is not installed. Please install it first.${NC}"
        exit 1
    fi
done

# Generate random string
generate_secret() {
    LC_ALL=C tr -dc 'A-Za-z0-9' </dev/urandom | head -c 32
}

# Detect if this is an update
IS_UPDATE=false
if [ -f .env ] && docker-compose ps -q &> /dev/null; then
    IS_UPDATE=true
    echo -e "${YELLOW}🔄 Detected existing deployment. Performing update...${NC}"

    # Pull latest changes from git
    if [ -d .git ]; then
        echo -e "${YELLOW}📥 Pulling latest code...${NC}"
        git pull origin main
    fi
else
    echo -e "${GREEN}🚀 Starting fresh deployment...${NC}"
fi

# Create .env if missing
if [ ! -f .env ]; then
    echo -e "${YELLOW}📄 .env not found. Creating from example...${NC}"
    if [ -f .env.example ]; then
        cp .env.example .env
        sed -i "s/WEBHOOK_SECRET=.*/WEBHOOK_SECRET=$(generate_secret)/" .env
        echo -e "${GREEN}✅ .env created with secure webhook secret${NC}"
    else
        echo -e "${RED}❌ .env.example is missing. Cannot continue.${NC}"
        exit 1
    fi
fi

# Load .env
set -a
source .env
set +a

# Build services
echo -e "${YELLOW}🔨 Building services...${NC}"
docker-compose build --no-cache

# Stop and remove old containers
echo -e "${YELLOW}🛑 Stopping existing containers...${NC}"
docker-compose down --remove-orphans

# Start new containers
echo -e "${YELLOW}🚀 Starting services...${NC}"
docker-compose up -d

# Show running containers
docker-compose ps

# Optional: Docker cleanup
echo -e "${YELLOW}🧹 Cleaning unused Docker resources...${NC}"

# Stop and remove stopped containers
docker container prune -f

# Remove unused images
docker image prune -af

# Remove unused networks
docker network prune -f

# Remove dangling volumes (does NOT touch your bind-mounted certs)
docker volume prune -f

# Clean up build cache
docker builder prune -af

# Final system-wide prune (excluding bind-mounts)
docker system prune -af --volumes

echo -e "${GREEN}✅ Deployment complete!${NC}"
echo -e "${YELLOW}📜 Use 'docker-compose logs -f' to follow logs.${NC}"
