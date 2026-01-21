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
```

## Docker

```bash
docker compose up --build
```

Check:

```bash
curl http://localhost:8080/api/v1/health/liveness
curl http://localhost:8080/api/v1/health/readiness
curl http://localhost:8080/api/v1/health
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
- internal/usecase - business logic
- internal/storage/postgres - Postgres repositories
- internal/storage/redis - Redis client and repositories
- internal/transport/httpserver - HTTP layer (Gin)
- internal/ratelimit - rate limiting primitives
- internal/logger - logging
