# ZeraBot 🤖

A Telegram bot built with Go that handles basic commands and messages using webhooks.

## Features

- `/start` - Welcome message
- `/help` - Show available commands
- `/hello` - Get a friendly greeting
- Webhook support with automatic HTTPS via Let's Encrypt

## Prerequisites

- Go 1.21 or higher
- A Telegram bot token from [@BotFather](https://t.me/botfather)
- A domain name (for webhook)
- Ports 80 and 443 open on your server

## Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/jfederk/ZeraBot.git
   cd ZeraBot
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Copy the example environment file and fill in your details:
   ```bash
   cp .env.example .env
   ```
   Then edit the `.env` file with your:
   - Telegram Bot Token
   - Domain name
   - Webhook secret (a long random string)
   - (Optional) Email for Let's Encrypt

4. Make sure your domain's DNS is pointing to your server's IP address.

## Running the Bot

1. Run the bot (requires root/sudo for ports 80/443):
   ```bash
   sudo -E go run main.go
   ```
   The `-E` flag preserves the environment variables.

2. The bot will:
   - Set up a webhook with Telegram
   - Obtain an SSL certificate from Let's Encrypt
   - Start listening for updates

3. Start a chat with your bot on Telegram and send `/start` to begin.

## Building for Production

1. Build the binary:
   ```bash
   go build -o zerabot
   ```

2. Run with systemd (recommended for production):
   ```
   [Unit]
   Description=ZeraBot Telegram Bot
   After=network.target

   [Service]
   Type=simple
   User=yourusername
   WorkingDirectory=/path/to/zerabot
   EnvironmentFile=/path/to/zerabot/.env
   ExecStart=/path/to/zerabot/zerabot
   Restart=always

   [Install]
   WantedBy=multi-user.target
   ```

## Security Notes

- Keep your `.env` file secure and never commit it to version control
- Use a strong, random string for `WEBHOOK_SECRET`
- Run the bot as a non-root user in production
- Keep your server and dependencies updated

## Contributing

Feel free to submit issues and enhancement requests.

## License

This project is open source and available under the [MIT License](LICENSE).
