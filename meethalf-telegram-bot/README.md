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

- /start - greet user by profile name when available, otherwise Telegram name; show inline Profile (or Create Profile when missing) and Settings buttons

## Profile setup

The bot guides users through a nine-step profile setup: a short anti-bot verification check, name (use the Telegram name
button or type a custom one), gender, birth date (YYYY-MM-DD), country (Russia, Kazakhstan, or Belarus), city (selected
from the supported list), description, emoji selection, and a photo album (1-4 photos). Gender, country, city, and emoji
are selected via inline buttons. After sending at least one photo, use the Done button to finish the setup. Drafts are
stored in the same session store so they can survive restarts when Redis is enabled. The setup header includes an estimated
total completion time (shown in minutes) calculated from per-step durations.

Supported cities:

- Russia: Moscow, Saint Petersburg, Novosibirsk, Krasnodar, Omsk, Rostov-on-Don, Perm, Krasnoyarsk, Yekaterinburg, Kazan,
  Nizhny Novgorod, Ufa, Chelyabinsk, Samara, Voronezh, Volgograd.
- Kazakhstan: Astana, Almaty, Semey, Pavlodar, Shymkent, Aktobe, Karaganda, Taraz, Ust-Kamenogorsk, Atyrau.
- Belarus: Minsk, Gomel, Mogilev, Vitebsk, Grodno, Brest, Bobruisk, Baranovichi, Borisov.

Supported profile emoji codes:

- LDR - 👑
- STR - 🧠
- ANA - 🧩
- CRT - 🎨
- COM - 🤝
- EMP - ❤️
- MED - 🕊
- PRF - 🧼
- RSR - 🧭
- INN - 💡
- EXE - 🛠
- ADV - 🔥
- CNT - ☕️
- RLS - 🧱
- MOT - 🎯
- SKP - 🛡

## Profile view

Use the inline `Profile` button to see the saved profile details. If the profile does not exist, the button is labeled
`Create Profile` and starts the setup flow. When a profile is missing and the user opens Profile or Preview, the bot
returns a not-found message and shows the `Create Profile` button instead of Preview/Edit actions. The response includes a
`Preview profile` button (shows how other users see the profile) above the `Edit profile` button to start editing. The
edit button opens a menu so users can choose which field to update (name, gender, birth date, country, city, description,
emoji, or photos). When the profile has photos, the bot sends them as an album with the profile details in the caption.
Profile creation and updates send a confirmation message followed by the full profile details with the `Edit profile`
button (including the number of saved photos).

## Loading messages

Operations that call the Meethalf API (profile view, edits, saves, and confirmed deletes) first send short, action-specific
loading messages so users see progress while the request is in flight. The loading message is deleted after the final
response is sent.

## Profile settings

Use the inline `Settings` button from `/start` to open profile settings. The menu includes a `Delete profile`
button that opens a confirmation step. If the profile does not exist, Settings returns a not-found message and shows the
`Create Profile` button instead of delete. Confirming removes the profile via the Meethalf API, and cancel keeps the
profile unchanged.

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
- BOT_ALLOWED_UPDATES (message,callback_query)
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
- API_BASE_URL (http://localhost:8080)
- API_TIMEOUT (5s)

`BOT_TOKEN` is required on startup. `.env` is loaded automatically if present. Long polling is used by
default. Use `SESSION_STORE=redis` with `REDIS_ENABLED=true` to share session state across multiple bot
instances.

## Structure

- cmd/bot - entrypoint
- internal/app - dependency wiring and run
- internal/config - config
- internal/domain - domain entities
- internal/usecase/bot - bot usecase (command routing, profile flow, response builders)
- internal/storage/memory - in-memory repositories
- internal/storage/redis - Redis repositories
- internal/transport/api - HTTP client for Meethalf API
- internal/transport/telegram - Telegram transport (poller, handler, sender)
- internal/logger - logging
