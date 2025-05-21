#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to generate a random string
generate_secret() {
    LC_ALL=C tr -dc 'A-Za-z0-9' </dev/urandom | head -c 32
}

# Function to check if services are healthy
check_services_health() {
    local timeout=60
    local start_time=$(date +%s)
    
    echo -e "${YELLOW}⏳ Checking services health...${NC}"
    
    while true; do
        local current_time=$(date +%s)
        local elapsed=$((current_time - start_time))
        
        if [ $elapsed -ge $timeout ]; then
            echo -e "${RED}❌ Services did not become healthy within $timeout seconds${NC}"
            return 1
        fi
        
        if docker-compose ps | grep -q "Up (healthy)"; then
            echo -e "${GREEN}✅ All services are healthy!${NC}"
            return 0
        elif docker-compose ps | grep -q "Exit"; then
            echo -e "${RED}❌ Some services have exited unexpectedly${NC}"
            docker-compose logs
            return 1
        fi
        
        sleep 5
    done
}

# Check for required commands
for cmd in docker docker-compose git; do
    if ! command_exists "$cmd"; then
        echo -e "${RED}❌ Error: $cmd is not installed. Please install it and try again.${NC}"
        exit 1
    fi
done

# Check if this is an update
IS_UPDATE=false
if [ -f .env ] && docker-compose ps -q &> /dev/null; then
    IS_UPDATE=true
    echo -e "${YELLOW}🔄 Detected existing deployment, performing update...${NC}"
    
    # Backup current version
    CURRENT_VERSION=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    echo -e "${YELLOW}📦 Backing up current version ($CURRENT_VERSION)...${NC}"
    
    # Pull latest changes
    if [ -d .git ]; then
        git pull origin main
    fi
else
    echo -e "${GREEN}🚀 Starting new deployment...${NC}"
    
    # Check if .env exists, if not create from example
    if [ ! -f .env ]; then
        echo -e "${YELLOW}ℹ️  .env file not found, creating from example...${NC}"
        if [ -f .env.example ]; then
            cp .env.example .env
            # Generate a secure webhook secret
            sed -i "s/WEBHOOK_SECRET=.*/WEBHOOK_SECRET=$(generate_secret)/" .env
            echo -e "${GREEN}✅ Created .env file with secure defaults${NC}"
        else
            echo -e "${RED}❌ Error: .env.example not found${NC}"
            exit 1
        fi
    fi
fi

# Load environment variables
set -a
source .env
set +a

# Build and start services
echo -e "${YELLOW}🔨 Building services...${NC}"
if ! docker-compose build --no-cache; then
    echo -e "${RED}❌ Failed to build Docker images${NC}"
    exit 1
fi

# Stop and remove existing containers if they exist
echo -e "${YELLOW}🛑 Stopping any running services...${NC}"
docker-compose down --remove-orphans || true

# Start the services in detached mode
echo -e "${YELLOW}🚀 Starting services...${NC}"
if ! docker-compose up -d; then
    echo -e "${RED}❌ Failed to start services${NC}"
    exit 1
fi

# Check services health
if ! check_services_health; then
    # Attempt to rollback if this was an update
    if [ "$IS_UPDATE" = true ]; then
        echo -e "${YELLOW}🔄 Attempting rollback...${NC}"
        git checkout $CURRENT_VERSION
        docker-compose up -d --force-recreate
        
        if check_services_health; then
            echo -e "${GREEN}✅ Successfully rolled back to previous version ($CURRENT_VERSION)${NC}"
        else
            echo -e "${RED}❌ Rollback failed. Manual intervention required.${NC}"
            exit 1
        fi
    else
        exit 1
    fi
fi

# Show deployment summary
echo -e "\n${GREEN}✅ Deployment successful!${NC}"
echo -e "\n📋 Deployment Summary:"
echo -e "  - Version: $(git rev-parse --short HEAD 2>/dev/null || echo "unknown")"
echo -e "  - Services: $(docker-compose ps --services | wc -l) services running"
echo -e "  - Mode: $([ "$IS_UPDATE" = true ] && echo "Update" || echo "Fresh Install")"
echo -e "\n🔍 Monitoring:"
echo -e "  - View logs: ${YELLOW}docker-compose logs -f${NC}"
echo -e "  - Check status: ${YELLOW}docker-compose ps${NC}"
echo -e "  - View bot info: ${YELLOW}curl -s http://localhost:8080/health | jq${NC} (requires jq)"

if [ "$IS_UPDATE" = false ]; then
    echo -e "\n🌐 Your bot should now be running! Try sending a message to it on Telegram."
fi
