# ZeraBot Deployment Guide

This guide explains how to deploy ZeraBot in a production environment using Docker.

## Prerequisites

1. Ubuntu 20.04/22.04 LTS server
2. Docker and Docker Compose installed
3. Domain name pointing to your server's IP
4. Ports 80, 443, and 8080 open in your firewall

## Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/ZeraBot.git
   cd ZeraBot
   ```

2. **Configure environment variables**
   ```bash
   cp .env.example .env
   nano .env
   ```
   
   Update the following variables:
   ```
   ENVIRONMENT=production
   DOMAIN=yourdomain.com
   TELEGRAM_BOT_TOKEN=your_telegram_bot_token
   WEBHOOK_SECRET=a_secure_random_string
   DATABASE_URL=postgresql://user:password@host:port/dbname?sslmode=require
   LETSENCRYPT_EMAIL=your_email@example.com
   ```

3. **Make the deployment script executable**
   ```bash
   chmod +x deploy.sh
   ```

## Deployment

1. **Run the deployment script**
   ```bash
   ./deploy.sh
   ```

   This will:
   - Build the Docker image
   - Start all services in detached mode
   - Set up automatic SSL certificates with Let's Encrypt

2. **Verify the deployment**
   ```bash
   docker-compose ps
   docker-compose logs -f
   ```

## Updating

To update to the latest version:

```bash
git pull origin main
docker-compose build --no-cache
docker-compose down
docker-compose up -d
```

## Backup and Restore

### Backup Database
```bash
docker exec -t zerabot-db pg_dump -U postgres zerabot > backup_$(date +%Y-%m-%d).sql
```

### Restore Database
```bash
cat backup_file.sql | docker exec -i zerabot-db psql -U postgres zerabot
```

## Monitoring

### View Logs
```bash
docker-compose logs -f
```

### Check Container Status
```bash
docker-compose ps
docker stats
```

## Security

1. **Firewall**
   ```bash
   sudo ufw allow 80/tcp
   sudo ufw allow 443/tcp
   sudo ufw allow 22/tcp
   sudo ufw enable
   ```

2. **Automatic Updates**
   Consider setting up unattended-upgrades:
   ```bash
   sudo apt install unattended-upgrades
   sudo dpkg-reconfigure -plow unattended-upgrades
   ```

## Troubleshooting

### SSL Certificate Issues
If Let's Encrypt fails to issue certificates:
1. Ensure your domain is correctly pointing to the server
2. Check port 80 is open and accessible
3. Check logs: `docker-compose logs -f`

### Database Connection Issues
1. Verify the database is running: `docker-compose ps`
2. Check connection string in `.env`
3. Check logs: `docker-compose logs db`

## Maintenance

### Prune Unused Docker Resources
```bash
docker system prune -a --volumes
```

### Update Docker Images
```bash
docker-compose pull
docker-compose up -d --force-recreate
```
