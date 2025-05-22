#!/bin/bash

# Check if running on Windows (Git Bash, WSL, etc.)
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" || "$OSTYPE" == "win32" ]]; then
    IS_WINDOWS=true
    # Enable case-insensitive pathname expansion on Windows
    shopt -s nocasematch
else
    IS_WINDOWS=false
fi

# Colors for output
if [ -t 1 ]; then  # Check if stdout is a terminal
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    NC='\033[0m' # No Color
else
    RED=''
    GREEN=''
    YELLOW=''
    NC=''
fi

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to run commands with appropriate sudo/doas
run_as_root() {
    if command_exists sudo; then
        sudo "$@"
    elif command_exists doas; then
        doas "$@"
    else
        echo -e "${RED}❌ Neither sudo nor doas found. Please run this script as root.${NC}" >&2
        exit 1
    fi
}

# Function to generate a random string
generate_secret() {
    LC_ALL=C tr -dc 'A-Za-z0-9' </dev/urandom | head -c 32
}

# Function to check if services are running (basic check)
check_services_running() {
    echo -e "${YELLOW}⏳ Starting services...${NC}"
    
    # Give services a moment to start
    sleep 5
    
    echo -e "${YELLOW}ℹ️  Checking service status...${NC}"
    docker-compose ps
    
    echo -e "\n${GREEN}✅ Deployment completed!${NC}"
    echo -e "${YELLOW}ℹ️  Use 'docker-compose logs -f' to view logs${NC}"
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

# Function to run docker-compose with the correct command
docker_compose() {
    if command_exists docker-compose; then
        docker-compose "$@"
    else
        docker compose "$@"
    fi
}

# Build the new service first to minimize downtime
echo -e "${YELLOW}🔨 Building new services...${NC}"
if ! docker_compose build --no-cache; then
    echo -e "${RED}❌ Failed to build services${NC}"
    exit 1
fi

# Only stop old services after new ones are built
echo -e "${YELLOW}🛑 Stopping old services...${NC}"
docker_compose down --remove-orphans || true

# Set up certs directory with proper permissions
echo -e "${YELLOW}🔧 Configuring certificate directory...${NC}"
CERT_DIR="./certs"

if [ "$IS_WINDOWS" = true ]; then
    # Windows specific commands
    if [ -d "$CERT_DIR" ]; then
        echo -e "${YELLOW}Removing existing certs directory...${NC}"
        rm -rf "$CERT_DIR"
    fi
    mkdir -p "$CERT_DIR"
    # On Windows, we can't easily set permissions like chmod/chown
else
    # Unix/Linux specific commands
    if [ -d "$CERT_DIR" ]; then
        echo -e "${YELLOW}Removing existing certs directory...${NC}"
        run_as_root rm -rf "$CERT_DIR"
    fi
    mkdir -p "$CERT_DIR"
    chmod 700 "$CERT_DIR"
    run_as_root chown -R 1000:1000 "$CERT_DIR"
fi

# Start the new services
echo -e "${YELLOW}🚀 Starting services...${NC}"
if ! docker_compose up -d; then
    echo -e "${RED}❌ Failed to start services${NC}"
    exit 1
fi

# Wait a moment for services to start
if [ "$IS_WINDOWS" = true ]; then
    timeout /t 5 /nobreak >nul 2>&1  # Windows sleep
else
    sleep 5  # Unix sleep
fi

# Show container status
echo -e "\n${YELLOW}📊 Container status:${NC}"
docker_compose ps

# Get the number of running services
SERVICE_COUNT=$(docker_compose ps --services 2>/dev/null | wc -l | tr -d '[:space:]')
# Ensure we have at least 1 service if the count is 0
[ "$SERVICE_COUNT" -eq 0 ] && SERVICE_COUNT=1

# Show deployment summary
echo -e "\n${GREEN}✅ Services restarted successfully!${NC}"
echo -e "\n📋 Status Summary:"
echo -e "  - Version: $(git rev-parse --short HEAD 2>/dev/null || echo \"unknown\")"
echo -e "  - Services: $SERVICE_COUNT services running"
echo -e "  - SSL Certs: ${GREEN}Configured in /app/certs${NC}"
echo -e "\n🔍 Useful Commands:"
echo -e "  - View logs: ${YELLOW}docker_compose logs -f${NC}"
echo -e "  - Check status: ${YELLOW}docker_compose ps${NC}"
echo -e "  - View bot info: ${YELLOW}curl -s http://localhost:8080/health | jq${NC} (requires jq)"

# Platform-specific instructions
if [ "$IS_WINDOWS" = true ]; then
    echo -e "  - Check SSL certs: ${YELLOW}dir certs${NC}"
    echo -e "\n${YELLOW}ℹ️  Windows Note:${NC} If you encounter permission issues with certificates, run this script as Administrator."
else
    echo -e "  - Check SSL certs: ${YELLOW}ls -la ./certs${NC}"
    echo -e "\n${YELLOW}ℹ️  Note:${NC} If you encounter permission issues, you may need to run: ${YELLOW}sudo chown -R 1000:1000 ./certs${NC}"
fi

if [ "$IS_UPDATE" = false ]; then
    echo -e "\n🌐 Your bot should now be running! Try sending a message to it on Telegram."
fi
