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
| **frontend** | 3000 | Next.js trading UI (3D landing, terminal) |
| **nginx** | 80 | API gateway |
| **order-service** | 8081 | Orders + in-process matching engine |
| **market-data** | 8082 | Tickers, order book, candles |
| **ledger-service** | 8083 | Double-entry ledger |
| **wallet-service** | 8084 | Deposits, withdrawals, balances |
| **blockchain-service** | 8085 | EVM isolation layer |
| **settlement-service** | 8096 | Trade settlement consumer |
| **notification-service** | 8097 | Event notifications |
| **auth-service** | 8090 | Stub (future) |
| **user-service** | 8091 | Stub (future) |
| **matching-service** | 8092 | Stub (future standalone engine) |
| **postgres** | 5432 | Persistence schema (future) |
| **redis** | 6379 | Cache, sessions, rate limits |
| **kafka** | 9092 | Event bus |
| **prometheus** | 9090 | Metrics |
| **grafana** | 3000 | Dashboards |

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

- Frontend + API: http://localhost
- Grafana: http://localhost:3000 (admin / admin)
- Prometheus: http://localhost:9090

### Infrastructure only

For local Go development with Docker infra:

```bash
docker compose -f docker/docker-compose.infra.yml up -d
```

Then run services locally:

```bash
KAFKA_BROKERS=localhost:9092 REDIS_ADDR=localhost:6379 go run ./order-service/cmd/order-service
```

## Build a single service

```bash
docker build -f docker/Dockerfile \
  --build-arg SERVICE_PATH=wallet-service/cmd/wallet-service \
  -t cex/wallet-service .
```

## API Gateway routes

| Path | Service |
|------|---------|
| `/api/v1/auth/` | auth-service |
| `/api/v1/users/` | user-service |
| `/api/v1/orders` | order-service |
| `/api/v1/matching/` | matching-service |
| `/api/v1/market/` | market-data |
| `/api/v1/ledger/` | ledger-service |
| `/api/v1/wallet/` | wallet-service |
| `/api/v1/blockchain/` | blockchain-service |

## Design notes

- **Matching runs off-chain** inside order-service (not on blockchain).
- **blockchain-service** is separate from matching-engine by design.
- **PostgreSQL** schema is pre-provisioned; services use in-memory stores until migrated.
- **auth**, **user**, **matching** stubs are placeholders for future services.
