# Docker Image Checker

Herramienta para verificar actualizaciones de imágenes Docker y enviar notificaciones via Telegram.

## Características

- ✅ Verificación de imágenes Docker actualizadas vs registros remotos
- 📱 Notificaciones via Telegram con plantillas personalizables
- 🔧 Configuración flexible (.env + YAML)
- 📊 Logs estructurados
- 🏗️ Arquitectura basada en patrones SOLID (Observer, Strategy)

## Estructura del Proyecto

```
docker-image-checker/
├── cmd/
│   └── checker/          # Aplicación principal
├── internal/
│   ├── config/          # Configuración
│   ├── docker/          # Cliente Docker y estrategias
│   ├── notification/    # Sistema de notificaciones (Observer)
│   └── model/          # Modelos de datos
├── pkg/                # Bibliotecas públicas
├── configs/            # Archivos de configuración por defecto
├── templates/          # Plantillas de notificación
└── logs/              # Directorio de logs
```

## Configuración

### .env
```env
TELEGRAM_BOT_TOKEN=your_bot_token_here
TELEGRAM_CHAT_ID=your_chat_id_here
DOCKER_HOST=unix:///var/run/docker.sock
LOG_LEVEL=info
```

### config.yaml
```yaml
checker:
  interval: "24h"
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

## Uso

```bash
# Ejecutar verificación única
./docker-image-checker

# Ejecutar en modo daemon
./docker-image-checker --daemon
```
