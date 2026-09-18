# CEX Docker Infrastructure

Containerized deployment for the full CEX platform.

## Architecture

```
                    ┌─────────────┐
                    │   Nginx     │  :80  API Gateway + Frontend
                    └──────┬──────┘
           ┌───────────────┼───────────────┐
           ▼               ▼               ▼
    order-service    market-data     wallet-service
    (matching)       ledger-service  blockchain-service
           │               │               │
           └───────────────┼───────────────┘
                           ▼
              ┌────────────┴────────────┐
              ▼                         ▼
           Kafka                      Redis
              │                         │
              ▼                         ▼
    settlement-service            PostgreSQL
    notification-service
              │
              ▼
    Prometheus ──► Grafana
```

| Container | Port | Role |
|-----------|------|------|
| **frontend** | 3002 | Next.js trading UI (Docker; local dev uses :3000) |
| **nginx** | 80 | API gateway |
| **order-service** | 8081 | Orders + in-process matching engine |
| **api-gateway** | 8080 | API front door (routing, CORS, rate limits) |
| **binance-adapter-service** | 8086 | Binance Spot adapter (mock without API keys) |
| **market-data** | 8082 | Tickers, order book, candles |
| **ledger-service** | 8083 | Double-entry ledger |
| **wallet-service** | 8084 | Deposits, withdrawals, balances |
| **blockchain-service** | 8085 | EVM isolation layer |
| **settlement-service** | 8096 | Trade settlement consumer |
| **notification-service** | 8097 | Event notifications |
| **auth-service** | 8090 | Register, login, JWT, refresh tokens, sessions |
| **user-service** | 8091 | User profiles (PostgreSQL) |
| **matching-service** | 8092 | Stub (future standalone engine) |
| **postgres** | internal only (full stack) / 5433 (infra) | PostgreSQL (users, ledger, wallet schemas) |
| **redis** | internal only (full stack) / 6380 (infra) | Cache, sessions, rate limits |
| **kafka** | 9092 | Event bus |
| **prometheus** | 9090 | Metrics |
| **grafana** | 3001 | Dashboards |

## Layout

```
docker/
├── docker-compose.yml          # Full stack
├── docker-compose.infra.yml    # Infrastructure only
├── Dockerfile                  # Multi-service Go builder
├── stubs/                      # Placeholder services
├── postgres/init/              # DB bootstrap schema
├── redis/redis.conf
├── nginx/                      # API gateway config
├── frontend/                   # Static frontend placeholder
└── monitoring/                 # Prometheus + Grafana
```

## Quick start

### Full stack

```bash
# From repo root
docker compose -f docker/docker-compose.yml up --build
```

- Frontend (Docker): http://localhost:3002
- Frontend (local dev): http://localhost:3000
- API via nginx: http://localhost:8088
- API gateway (direct): http://localhost:8080
- Grafana: http://localhost:3001 (admin / admin)
- Prometheus: http://localhost:9090

### Infrastructure only

For local Go development with Docker infra:

```bash
docker compose -f docker/docker-compose.infra.yml up -d
```

Then run services locally:

```bash
KAFKA_BROKERS=localhost:9092 REDIS_ADDR=localhost:6380 go run ./order-service/cmd/order-service
```

## Build a single service

```bash
docker build -f docker/Dockerfile \
  --build-arg SERVICE_PATH=wallet-service/cmd/wallet-service \
  -t cex/wallet-service .
```

## API routing

Nginx (port 80) proxies `/api/*` to **api-gateway** (port 8080), which routes to internal services.

| Gateway path | Service |
|--------------|---------|
| `/api/auth`, `/api/v1/auth` | auth-service |
| `/api/users`, `/api/v1/users` | user-service |
| `/api/orders`, `/api/v1/orders` | order-service |
| `/api/matching`, `/api/v1/matching` | matching-service |
| `/api/market`, `/api/v1/market` | market-data |
| `/api/ledger`, `/api/v1/ledger` | ledger-service |
| `/api/wallet`, `/api/v1/wallet` | wallet-service |
| `/api/blockchain`, `/api/v1/blockchain` | blockchain-service |
| `/api/admin`, `/api/v1/admin` | admin-service |

## Design notes

- **Matching runs off-chain** inside order-service (not on blockchain).
- **blockchain-service** is separate from matching-engine by design.
- **PostgreSQL** schema is pre-provisioned; services use in-memory stores until migrated.
- **auth**, **user**, **matching** stubs are placeholders for future services.
