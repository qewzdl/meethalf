# meethalf-telegram-bot

Base Telegram bot scaffold on Go with clean architecture layering.

## Run

Create `.env` (or copy `.env.example`) and set your token:

```bash
BOT_TOKEN=telegram-bot-token
```

Start:

```bash
go run ./cmd/bot
```

## Commands

- /start - greet user by first and last name, show help
- /help - show help
- /ping - health check

## Docker

```bash
docker compose up --build
```

Stop:

```bash
docker compose down
```

## Environment

- APP_ENV (dev)
- BOT_TOKEN ()
- BOT_DEBUG (false)
- BOT_ALLOWED_UPDATES (message)
- BOT_POLLING_TIMEOUT (10s)
- BOT_WORKERS (4)
- BOT_QUEUE_SIZE (100)
- SESSION_STORE (memory)
- SESSION_TTL (24h)
- REDIS_ENABLED (false)
- REDIS_HOST (localhost)
- REDIS_PORT (6379)
- REDIS_USERNAME ()
- REDIS_PASSWORD ()
- REDIS_DB (0)
- REDIS_CONNECT_TIMEOUT (5s)
- REDIS_READ_TIMEOUT (3s)
- REDIS_WRITE_TIMEOUT (3s)
- REDIS_POOL_SIZE (0)
- REDIS_MIN_IDLE_CONNS (0)

`BOT_TOKEN` is required on startup. `.env` is loaded automatically if present. Long polling is used by
default. Use `SESSION_STORE=redis` with `REDIS_ENABLED=true` to share session state across multiple bot
instances.

## Structure

- cmd/bot - entrypoint
- internal/app - dependency wiring and run
- internal/config - config
- internal/domain - domain entities
- internal/usecase/bot - business logic
- internal/storage/memory - in-memory session repository
- internal/storage/redis - Redis session repository
- internal/transport/telegram - Telegram transport (poller, handler, sender)
- internal/logger - logging
