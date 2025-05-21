# ZeraBot 🤖

ZeraBot is a high-performance Telegram bot that monitors Zera Network governance proposals and delivers real-time updates to subscribers. Built with Go and designed for reliability, it supports thousands of concurrent users while maintaining minimal resource usage.

## ✨ Core Features

- **Real-time Proposal Tracking**: Monitor new governance proposals as they're created
- **Symbol-based Subscriptions**: Users can subscribe to specific proposals using symbols (e.g., `$ZRA+0000`)
- **Group Management**: Admins can manage subscriptions for their communities

## 🚀 End-to-End Deployment

### Prerequisites
- Linux server (Ubuntu 22.04 recommended)
- Docker and Docker Compose
- Domain name pointing to your server
- Ports 22, 80, 443, and 8080 open

### 1. Server Setup (First-Time Only)

```bash
# Update system and install requirements
sudo apt update && sudo apt upgrade -y
sudo apt install -y git curl jq ufw

# Configure firewall
sudo ufw allow OpenSSH
sudo ufw allow http
sudo ufw allow https
sudo ufw allow 8080/tcp
sudo ufw allow 50051/tcp
sudo ufw --force enable

# Install Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

### 2. Deploy ZeraBot

```bash
# Clone the repository
sudo git clone https://github.com/jfederk/ZeraBot.git /opt/zerabot
cd /opt/zerabot

# Create and configure .env file
cp .env.example .env
nano .env  # Edit with your details

# Make deployment script executable
chmod +x deploy.sh

# Run the deployment
./deploy.sh
```

### 3. Required .env Configuration

```env
# Telegram Bot Token from @BotFather
TELEGRAM_BOT_TOKEN=your_telegram_bot_token_here

# Your domain (e.g., bot.yourdomain.com)
DOMAIN=your_domain_here

# A secret key for your webhook endpoint (use a long, random string)
WEBHOOK_SECRET=your_webhook_secret_here

# Email for Let's Encrypt (optional but recommended)
LETSENCRYPT_EMAIL=your_email@example.com

# Set to "development" for local development, "production" for live
ENVIRONMENT=production
#NGROK_URL=https://your-ngrok-url.ngrok.io  # Replace with your actual ngrok URL (if using for localdev)

# Database connection string (PostgreSQL) #! VPC Network preferred
DATABASE_URL=postgres://username:password@localhost:5432/zerabot?sslmode=disable

# Expected gossip address (zera network address (expects domain, ie routin.zera.vision - but can be modified to accept ipv4))
GRPC_ADDRESS=domain.example.com

```

### 4. Verify Installation

```bash
# Check running services
docker-compose ps

# View logs
docker-compose logs -f

# Check health status
curl -s http://localhost:8080/health | jq
```

## 🔄 Updating ZeraBot

```bash
# Navigate to installation directory
cd /opt/zerabot

# Pull latest changes
sudo git pull (force as needed: git fetch origin && git reset --hard origin/main && chmod +x deploy.sh)

# Run deployment
./deploy.sh
```

The deployment script will automatically:
- Pull the latest code changes
- Rebuild containers if needed
- Restart services with zero downtime

## 🛠️ Maintenance

### Common Commands

```bash
# View logs
docker-compose logs -f

# Restart services
docker-compose restart
```

## 🤖 Bot Commands

- `/start` - Welcome message and brief introduction
- `/help` - Show available commands
- `/proposalSubscribe $SYMBOL` - Subscribe to a specific proposal symbol
- `/proposalUnsubscribe $SYMBOL` - Unsubscribe from a specific proposal
- `/proposalSubscribe all` - Subscribe to all proposals (admin only in groups)
- `/proposalUnsubscribe all` - Unsubscribe from all proposals (admin only in groups)
- `/mysubscriptions` - List your current subscriptions

## 🐳 Docker Deployment

### Development

```bash
docker-compose -f docker-compose.dev.yml up --build
```

### Production

```bash
# Build and start services
docker-compose up -d --build

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

## 📊 Database

ZeraBot uses PostgreSQL for storing subscriptions. The database schema is managed through migrations.

## 🔒 Security

- All sensitive data is stored in environment variables
- HTTPS is enforced for all webhook communications
- Group admin verification for sensitive commands

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request