# 🐳 Docker Image Checker

Tool to check Docker image updates and send notifications via Telegram.

## ✨ Features

- ✅ Docker image verification against remote registries
- 📱 Telegram notifications with customizable templates
- 🔧 Flexible configuration (.env + YAML)
- 📊 Structured logging
- 🏗️ Architecture based on SOLID patterns (Observer, Strategy)

## ⚙️ Configuration

### 🔐 .env
```env
TELEGRAM_BOT_TOKEN=your_bot_token_here
TELEGRAM_CHAT_ID=your_chat_id_here
DOCKER_HOST=unix:///var/run/docker.sock
LOG_LEVEL=info
```

### 📋 config.yaml
```yaml
checker:
  schedule: "0 0 * * *"  # Cron format: daily at midnight
  exclude_images:
    - "local/custom-image"
    - "build-*"
  include_build_images: false

notifications:
  telegram:
    enabled: true
    template_file: "templates/telegram-template.tmpl"

logging:
  file: "logs/checker.log"
  max_size: 10
  max_backups: 3
```

## 🚀 Usage

```bash
# 🔍 Run single check
./docker-image-checker

# 🔍 Run single check and exit
./docker-image-checker --once

# 🤖 Run in daemon mode with cron schedule
./docker-image-checker --daemon
```

## ⏰ Schedule Configuration

The `schedule` field uses standard cron format:

```
┌───────────── minute (0 - 59)
│ ┌───────────── hour (0 - 23)
│ │ ┌───────────── day of month (1 - 31)
│ │ │ ┌───────────── month (1 - 12)
│ │ │ │ ┌───────────── day of week (0 - 6, Sunday = 0)
│ │ │ │ │
* * * * *
```

📅 **Examples:**
- `"0 0 * * *"` - 🌙 Daily at midnight
- `"0 12 * * *"` - ☀️ Daily at 12:00 PM
- `"0 */6 * * *"` - ⏰ Every 6 hours
- `"0 9 * * 1"` - 📅 Mondays at 9:00 AM
