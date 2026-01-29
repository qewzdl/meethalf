# meethalf-api

Base API scaffold on Go + Gin with clean architecture layering.

## Run

```bash
go run ./cmd/api
```

Check:

```bash
curl http://localhost:8080/api/v1/health/liveness
curl http://localhost:8080/api/v1/health/readiness
curl http://localhost:8080/api/v1/health
curl -X POST http://localhost:8080/api/v1/profiles \
  -H "Content-Type: application/json" \
  -d '{"user_id":1,"username":"janedoe","name":"Jane Doe","gender":"female","birth_date":"1996-04-23","country":"russia","city":"Moscow","description":"Hello from Meethalf","emoji_code":"LDR","photos":["photo-1","photo-2"],"is_hidden":false}'
curl http://localhost:8080/api/v1/profiles/1
curl -X PATCH http://localhost:8080/api/v1/profiles/1/visibility \
  -H "Content-Type: application/json" \
  -d '{"is_hidden":true}'
curl -X DELETE http://localhost:8080/api/v1/profiles/1
curl -X POST http://localhost:8080/api/v1/search/start \
  -H "Content-Type: application/json" \
  -d '{"user_id":1,"gender":"female","accuracy":3}'
curl -X POST http://localhost:8080/api/v1/search/next \
  -H "Content-Type: application/json" \
  -d '{"user_id":1}'
curl -X POST http://localhost:8080/api/v1/search/previous \
  -H "Content-Type: application/json" \
  -d '{"user_id":1}'
curl -X POST http://localhost:8080/api/v1/search/action \
  -H "Content-Type: application/json" \
  -d '{"user_id":1,"target_id":2,"action":"like"}'
# {"matched":false}
curl http://localhost:8080/api/v1/search/history/1?limit=20&offset=0
curl http://localhost:8080/api/v1/likes/1
curl http://localhost:8080/api/v1/admin/users?limit=20&offset=0
curl http://localhost:8080/api/v1/admin/users/1
curl http://localhost:8080/api/v1/admin/users?limit=20&offset=0&banned=true
curl http://localhost:8080/api/v1/admin/users?limit=20&offset=0&moderator=true
curl http://localhost:8080/api/v1/admin/reports?limit=20&offset=0
curl -X POST http://localhost:8080/api/v1/admin/users/1/ban
curl -X POST http://localhost:8080/api/v1/admin/users/1/unban
curl -X POST http://localhost:8080/api/v1/admin/users/1/moderator
curl -X POST http://localhost:8080/api/v1/admin/users/1/unmoderator
curl -X POST http://localhost:8080/api/v1/admin/users/1/reports/clear
curl -X POST http://localhost:8080/api/v1/admin/users/1/choices/reset
```

`birth_date` uses the `YYYY-MM-DD` format; age is derived automatically and must be between 16 and 120. `country` must be one of `russia`, `kazakhstan`,
or `belarus`; `city` must be in the supported list for the selected country. `emoji_code` must be one of the supported
profile emoji codes listed below. Set `is_hidden=true` to hide a profile from search results.
Profile responses include `is_moderator` to indicate moderation role; it is managed through the admin endpoints.
Search and likes endpoints require an existing profile. `gender` can be `male`, `female`, `other`, or `unspecified` (any),
and `accuracy` is a 0-4 scale where 0 is wider/random and 4 is stricter. If no candidates match the selected accuracy,
search relaxes the accuracy step-by-step down to 0 while keeping the gender filter. Lower accuracy levels also use wider
age windows when scoring candidates.
`/api/v1/search/action` responds with `matched=true` when the like action forms a mutual match.
`GET /api/v1/search/history/{user_id}` returns the cumulative search history across sessions (latest view per profile,
latest first) with actions, position, and pagination via `limit`/`offset` query parameters.
`/api/v1/admin/users` returns a paginated list of users (profiles) with `username`, `is_hidden`, `is_banned`, and
`is_moderator` flags.
`GET /api/v1/admin/users/{user_ref}` returns a single user summary by id or username (with or without `@`).
Use `limit` and `offset` query parameters to paginate; pass `banned=true` to list only banned users or
`moderator=true` to list only moderators. `/api/v1/admin/reports` returns a paginated list of reported users with
their report counts.
`POST /api/v1/admin/users/{user_ref}/ban` bans the user by id or username (with or without `@`), and
`POST /api/v1/admin/users/{user_ref}/unban` removes the ban. `POST /api/v1/admin/users/{user_ref}/moderator` assigns the
moderator role, and `POST /api/v1/admin/users/{user_ref}/unmoderator` removes it.
`POST /api/v1/admin/users/{user_ref}/reports/clear` removes the user from the reported list by clearing report actions.
`POST /api/v1/admin/users/{user_ref}/choices/reset` clears all match choices and history for the selected user.
Banned users cannot use profile or search endpoints.

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

## Docker

```bash
docker compose up --build
```

Check:

```bash
curl http://localhost:8080/api/v1/health/liveness
curl http://localhost:8080/api/v1/health/readiness
curl http://localhost:8080/api/v1/health
curl -X POST http://localhost:8080/api/v1/profiles \
  -H "Content-Type: application/json" \
  -d '{"user_id":1,"username":"janedoe","name":"Jane Doe","gender":"female","birth_date":"1996-04-23","country":"russia","city":"Moscow","description":"Hello from Meethalf","emoji_code":"LDR","photos":["photo-1","photo-2"],"is_hidden":false}'
curl http://localhost:8080/api/v1/profiles/1
curl -X PATCH http://localhost:8080/api/v1/profiles/1/visibility \
  -H "Content-Type: application/json" \
  -d '{"is_hidden":true}'
curl -X DELETE http://localhost:8080/api/v1/profiles/1
curl -X POST http://localhost:8080/api/v1/search/start \
  -H "Content-Type: application/json" \
  -d '{"user_id":1,"gender":"female","accuracy":3}'
curl -X POST http://localhost:8080/api/v1/search/next \
  -H "Content-Type: application/json" \
  -d '{"user_id":1}'
curl -X POST http://localhost:8080/api/v1/search/previous \
  -H "Content-Type: application/json" \
  -d '{"user_id":1}'
curl -X POST http://localhost:8080/api/v1/search/action \
  -H "Content-Type: application/json" \
  -d '{"user_id":1,"target_id":2,"action":"like"}'
# {"matched":false}
curl http://localhost:8080/api/v1/search/history/1?limit=20&offset=0
curl http://localhost:8080/api/v1/likes/1
curl http://localhost:8080/api/v1/admin/users?limit=20&offset=0
curl http://localhost:8080/api/v1/admin/users/1
curl http://localhost:8080/api/v1/admin/users?limit=20&offset=0&banned=true
curl http://localhost:8080/api/v1/admin/users?limit=20&offset=0&moderator=true
curl http://localhost:8080/api/v1/admin/reports?limit=20&offset=0
curl -X POST http://localhost:8080/api/v1/admin/users/1/ban
curl -X POST http://localhost:8080/api/v1/admin/users/1/unban
curl -X POST http://localhost:8080/api/v1/admin/users/1/moderator
curl -X POST http://localhost:8080/api/v1/admin/users/1/unmoderator
curl -X POST http://localhost:8080/api/v1/admin/users/1/reports/clear
curl -X POST http://localhost:8080/api/v1/admin/users/1/choices/reset
```

Stop:

```bash
docker compose down
```

## Database init

```bash
go run ./cmd/db
```

The command connects to `DB_ADMIN_NAME` using `DB_ADMIN_USER`/`DB_ADMIN_PASSWORD`, creates or updates the app role,
and ensures `DB_NAME` exists with the app role as the owner. It hardens the database by revoking public access on the
database and schema, granting access only to `DB_USER`, setting the app role `search_path` to `DB_SCHEMA`, and revoking
default PUBLIC privileges for new objects in the schema.
Defaults match docker-compose so you can run without a `.env` file in dev.
For production, override the default passwords with strong values.
To keep using the public schema, set `DB_SCHEMA=public`.

## Environment

- APP_ENV (dev)
- HTTP_HOST (0.0.0.0)
- HTTP_PORT (8080)
- HTTP_READ_TIMEOUT (5s)
- HTTP_WRITE_TIMEOUT (10s)
- HTTP_IDLE_TIMEOUT (30s)
- HTTP_SHUTDOWN_TIMEOUT (10s)
- RATE_LIMIT_ENABLED (true)
- RATE_LIMIT_STORE (memory)
- RATE_LIMIT_REQUESTS (60)
- RATE_LIMIT_WINDOW (1m)
- RATE_LIMIT_BURST (60)
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
- HEALTH_DB_TIMEOUT (5s)
- HEALTH_REDIS_TIMEOUT (3s)
- DB_HOST (localhost)
- DB_PORT (5432)
- DB_USER (meethalf_app)
- DB_PASSWORD (meethalf_app)
- DB_NAME (meethalf)
- DB_SCHEMA (meethalf)
- DB_ADMIN_NAME (postgres)
- DB_ADMIN_USER (postgres)
- DB_ADMIN_PASSWORD (postgres)
- DB_SSLMODE (disable)
- DB_CONNECT_TIMEOUT (5s)
- DB_MAX_OPEN_CONNS (25)
- DB_MAX_IDLE_CONNS (25)
- DB_CONN_MAX_LIFETIME (30m)
- DB_CONN_MAX_IDLE_TIME (5m)

API requires a reachable PostgreSQL instance on startup. If `REDIS_ENABLED=true` (required for
`RATE_LIMIT_STORE=redis`), it also requires Redis. `/api/v1/health/readiness` checks the database and Redis (when enabled).
`/api/v1/health` returns dependency details with timeouts, while `/api/v1/health/liveness` only checks the process.
Rate limiting uses a token bucket per client IP; set `RATE_LIMIT_STORE=redis` to share limits across instances.

## Structure

- cmd/api - entrypoint
- cmd/db - database bootstrap command
- internal/app - dependency wiring and run
- internal/config - config
- internal/domain - domain entities
- internal/usecase/database - database provisioning
- internal/usecase/matching - matching and interactions logic
- internal/usecase/profile - profile logic
- internal/usecase - business logic
- internal/storage/postgres - Postgres repositories
- internal/storage/redis - Redis client and repositories
- internal/transport/httpserver - HTTP layer (Gin)
- internal/ratelimit - rate limiting primitives
- internal/logger - logging
