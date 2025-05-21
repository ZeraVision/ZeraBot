# ZeraBot 🤖

A Telegram bot built with Go that tracks Zera Network governance proposals and notifies subscribers about new proposals and updates. Built with scalability and security in mind.

## ✨ Features

- **Real-time Notifications**: Get instant updates about new governance proposals
- **Symbol-based Subscriptions**: Subscribe to specific proposal symbols (e.g., `$ZRA+0000`)
- **Admin Controls**: Group admins can manage subscriptions for their groups
- **Docker Support**: Easy deployment with Docker and Docker Compose
- **Secure**: Built with security best practices in mind

## 🚀 Quick Start with Docker

The easiest way to run ZeraBot is using Docker Compose:

```bash
# 1. Clone the repository
git clone https://github.com/yourusername/ZeraBot.git
cd ZeraBot

# 2. Configure environment variables
cp .env.example .env
nano .env  # Update with your configuration

# 3. Start the services
docker-compose up -d
```

## 📋 Prerequisites

- Docker and Docker Compose
- A Telegram bot token from [@BotFather](https://t.me/botfather)
- A domain name pointing to your server (for webhooks)
- Ports 80, 443, and 8080 open on your server

## 🔧 Configuration

Edit the `.env` file with your configuration:

```env
# Application
ENVIRONMENT=production
DOMAIN=yourdomain.com

# Telegram
TELEGRAM_BOT_TOKEN=your_telegram_bot_token

# Security
WEBHOOK_SECRET=a_secure_random_string

# Database
DATABASE_URL=postgresql://user:password@host:port/dbname?sslmode=require

# Let's Encrypt (optional)
LETSENCRYPT_EMAIL=your_email@example.com
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

ZeraBot uses PostgreSQL for storing subscriptions. The database schema is automatically managed through migrations.

### Backup

```bash
docker-compose exec db pg_dump -U postgres zerabot > backup_$(date +%Y-%m-%d).sql
```

### Restore

```bash
cat backup_file.sql | docker-compose exec -T db psql -U postgres zerabot
```

## 🔒 Security

- All sensitive data is stored in environment variables
- HTTPS is enforced for all webhook communications
- Group admin verification for sensitive commands
- Regular security updates for all dependencies

## 📚 Documentation

For detailed deployment and configuration instructions, see the [DEPLOYMENT.md](DEPLOYMENT.md) file.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Zera Network](https://zera.vision) for the amazing platform
- All the open-source projects that made this possible
