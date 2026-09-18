# CEX — Centralized Exchange

A Go microservices monorepo for building a centralized cryptocurrency exchange. The **matching engine** is the first implemented component.

## Matching Engine

Location: `matching-engine/`

### Concepts implemented

| Concept | Implementation |
|---------|----------------|
| Limit orders | Rest on book when not fully filled |
| Market orders | Walk the book at best available prices |
| Order book | Price levels with FIFO queues (price-time priority) |
| Partial fills | `Remaining` / `Filled` tracking per order |
| Cancel orders | Remove from book + unlock funds |
| Trade execution | Match at **maker's price** |
| Maker/taker | Fees computed per role (configurable bps) |
| Sequence numbers | Monotonic order and trade sequences |
| Atomic settlement | Ledger locks on submit, settles per trade |

### Fixed-point math

All amounts use `int64` (no floats):

- **Price**: whole USDT per 1 BTC (e.g. `101100`)
- **Quantity**: base asset with scale `100` (e.g. `30` = 0.30 BTC)
- **Notional**: `price * quantity / QuantityScale`

### Run the demo

```bash
go run ./matching-engine/cmd/demo
```

### Run tests

```bash
go test ./matching-engine/tests/ -v
```

### Example flow (BTC/USDT)

```
BUY 101100 (0.30 BTC)  →  matches SELL 101100 (maker)
TRADE: 0.30 BTC @ 101100
Buyer:  USDT -30,330   BTC +0.30
Seller: BTC  -0.30     USDT +30,330
```

## API Gateway

Location: `api-gateway/`

The **front door** of the CEX. The frontend talks to one gateway instead of calling each microservice directly.

```
Frontend
    │
    ▼
API Gateway (:8080)
    │
    ├── /api/auth      → auth-service
    ├── /api/users     → user-service
    ├── /api/orders    → order-service
    ├── /api/wallet    → wallet-service
    ├── /api/market    → market-data
    ├── /api/admin     → admin-service
    └── /api/v1/*      → same services (versioned paths)
```

Production edge stack:

```
Internet → Nginx (TLS) → API Gateway → internal services
```

### Gateway responsibilities

| Concern | Implementation |
|---------|----------------|
| Routing | Path-based reverse proxy to all CEX services |
| API versioning | `/api/v1/*` passthrough + shorter `/api/*` aliases |
| Auth forwarding | `Authorization`, `X-Admin-API-Key` |
| Rate limiting | Redis (or in-memory fallback) |
| CORS | Configurable allowed origins |
| Request logging | Structured access logs with `X-Request-ID` |
| TLS | Optional via `TLS_CERT` / `TLS_KEY` on gateway |

### Examples

```bash
# Short path (rewritten to /api/v1/orders on order-service)
curl -X POST http://localhost:8080/api/orders \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"symbol":"BTC/USDT","side":"BUY","type":"LIMIT","price":101100,"quantity":30}'

# Versioned path (used by frontend today)
curl http://localhost:8080/api/v1/market/tickers

# Via Nginx edge (port 80)
curl http://localhost/api/v1/auth/login ...
```

### Run api-gateway

```bash
ORDER_SERVICE_URL=http://localhost:8081 \
AUTH_SERVICE_URL=http://localhost:8090 \
USER_SERVICE_URL=http://localhost:8091 \
go run ./api-gateway/cmd/api-gateway
```

Env vars: `*_SERVICE_URL` for each backend, `REDIS_ADDR`, `RATE_LIMIT_PER_MINUTE`, `CORS_ALLOWED_ORIGINS`, `TLS_CERT`, `TLS_KEY`.

### Run gateway tests

```bash
go test ./api-gateway/tests/ -v
```

## Auth Service

Location: `auth-service/`

Handles registration, login, password hashing (bcrypt), JWT access tokens, refresh tokens, logout, and session management. **2FA is reserved for a later phase.**

### Login flow

```
Frontend
   │  email + password
   ▼
API Gateway
   ▼
Auth Service
   ├── Verify password (bcrypt)
   ├── Create access token (JWT)
   └── Create refresh token (opaque session)
   ▼
Frontend stores tokens
```

On **register**, auth-service creates the user profile in user-service (PostgreSQL) first, then stores credentials locally:

```
Auth Service → User Service → PostgreSQL
```

Protected requests send:

```
Authorization: Bearer <JWT>
```

Access tokens identify the user only (`sub`, `role`, `exp`). They are **not** the source of truth for wallet balance, order status, or trade history — those belong to ledger, wallet, and order services.

Example JWT payload:

```json
{
  "sub": "user-123",
  "role": "TRADER",
  "exp": 1780000000
}
```

### HTTP API (port 8090)

| Endpoint | Description |
|----------|-------------|
| `POST /api/v1/auth/register` | Create account + issue tokens |
| `POST /api/v1/auth/login` | Verify credentials + issue tokens |
| `POST /api/v1/auth/refresh` | Rotate refresh token + new access token |
| `POST /api/v1/auth/logout` | Revoke refresh session |
| `GET /api/v1/auth/me` | Current user from bearer JWT |

```bash
# Register
curl -X POST http://localhost:8090/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"trader@cex.test","password":"password123"}'

# Login
curl -X POST http://localhost:8090/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"trader@cex.test","password":"password123"}'

# Use access token on protected routes (e.g. order-service with JWT_SECRET set)
curl http://localhost:8090/api/v1/auth/me \
  -H "Authorization: Bearer <access_token>"
```

Via API gateway: `http://localhost/api/v1/auth/...`

### Run auth-service

```bash
JWT_SECRET=cex-dev-jwt-secret-change-in-production go run ./auth-service/cmd/auth-service
```

Env vars: `JWT_SECRET`, `JWT_ISSUER`, `ACCESS_TOKEN_TTL` (default `15m`), `REFRESH_TOKEN_TTL` (default `168h`).

### JWT integration (order-service)

When `JWT_SECRET` is set on order-service, bearer tokens from auth-service are validated via shared `pkg/jwt`:

```bash
JWT_SECRET=cex-dev-jwt-secret-change-in-production go run ./order-service/cmd/order-service
```

### Run auth tests

```bash
go test ./auth-service/tests/ ./pkg/jwt/ -v
```

## User Service

Location: `user-service/`

Stores user profile information in **PostgreSQL**. Credentials stay in auth-service; user-service is the source of truth for profile fields.

```
User
├── ID
├── email
├── username
├── status
├── KYC status
├── created_at
└── updated_at
```

### Flow

```
Auth Service
      │  POST /api/v1/users (on register)
      ▼
User Service
      │
      ▼
PostgreSQL (users.accounts)
```

### HTTP API (port 8091)

| Endpoint | Description |
|----------|-------------|
| `POST /api/v1/users` | Create user profile |
| `GET /api/v1/users/:id` | Get user by ID |
| `GET /api/v1/users/by-email/:email` | Get user by email |
| `PATCH /api/v1/users/:id` | Update username, status, kyc_status |

```bash
# Create profile (normally called by auth-service)
curl -X POST http://localhost:8091/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email":"trader@cex.test","username":"trader1"}'

# Get user
curl http://localhost:8091/api/v1/users/user-abc123

# Update KYC status
curl -X PATCH http://localhost:8091/api/v1/users/user-abc123 \
  -H "Content-Type: application/json" \
  -d '{"kyc_status":"APPROVED"}'
```

### Run user-service

```bash
DATABASE_URL=postgres://cex:cex@localhost:5433/cex?sslmode=disable \
go run ./user-service/cmd/user-service
```

Start PostgreSQL first (via Docker infra or full stack).

### Run user tests

```bash
go test ./user-service/tests/ -v
```

## Order Service

Location: `order-service/`

### Request pipeline

```
Client Request
      │
      ▼
Authenticate user        (middleware/auth)
      │
      ▼
Validate symbol          (validator/symbol)
      │
      ▼
Validate price           (validator/price)
      │
      ▼
Validate quantity        (validator/quantity)
      │
      ▼
Check trading rules      (validator/trading_rules)
      │
      ▼
Check available balance  (validator/balance)
      │
      ▼
Matching Engine          (client → matching-engine/pkg/api)
```

### Run the order service

```bash
go run ./order-service/cmd/order-service
```

### Place an order (HTTP)

```bash
curl -X POST http://localhost:8081/api/v1/orders \
  -H "Authorization: Bearer token-user-a" \
  -H "Content-Type: application/json" \
  -d '{"symbol":"BTC/USDT","side":"BUY","type":"LIMIT","price":101100,"quantity":30}'
```

### Run order-service tests

```bash
go test ./order-service/tests/ -v
```

## Redis Layer

Redis is **not** on the matching engine critical path. It supports fast temporary/shared state around the exchange.

```
Order → Matching Engine (RAM) → Match     ✅ correct path
Order → Redis → Matching Engine             ❌ NOT used
```

### Redis use cases

| Use case | Service | Key pattern |
|----------|---------|-------------|
| Market-data cache | market-data | `cex:market:ticker:BTC-USDT` |
| Session cache | order-service | `cex:session:{token}` |
| Rate limits | order-service | `cex:ratelimit:{user}:{path}` |
| Temporary order state | order-service | `cex:order:pending:{id}` |
| Distributed locks | pkg/redis | `cex:lock:{resource}` |
| Pub/Sub (ticker feed) | market-data | `cex:pubsub:ticker:BTC-USDT` |

### Ticker cache flow

```
GET /api/ticker/BTC-USDT
        │
        ▼
      Redis  (cache hit → response)
        │
     miss ▼
   In-memory store → populate Redis → response
```

### Start Redis

```bash
docker compose -f docker/docker-compose.yml up -d redis
REDIS_ADDR=localhost:6379 go run ./market-data/cmd/market-data
REDIS_ADDR=localhost:6379 go run ./order-service/cmd/order-service
```

### Run Redis tests

```bash
go test ./pkg/redis/tests/ ./market-data/tests/ -v
```

## Kafka Event System

```
Matching Engine
       │
       ▼ publishes: orders, trades, orderbook
     Kafka
       │
 ┌─────┼────────┬──────────┐
 ▼     ▼        ▼          ▼
Market Ledger Settlement Notification
Data
```

### Topics

| Topic | Publisher | Consumers |
|-------|-----------|-----------|
| `orders` | Matching Engine | Notification |
| `trades` | Matching Engine | Market Data, Ledger, Settlement, Notification |
| `orderbook` | Matching Engine | Market Data |
| `ledger` | Ledger Service | (downstream accounting) |
| `settlement` | Settlement Service | Notification |
| `notifications` | Notification Service | (email/push workers) |

### Shared packages

- `pkg/events/` — event types, topics, envelope
- `pkg/kafka/` — Kafka producer, consumer, in-memory bus

### Start Kafka (Docker)

```bash
docker compose -f docker/docker-compose.yml up -d
```

### Market Data Service

Location: `market-data/`

Consumes `trades` and `orderbook` Kafka events and produces:

| Output | Description |
|--------|-------------|
| **Ticker** | Last price, bid/ask, 24h change |
| **Order Book** | Live bids and asks depth |
| **Trades** | Recent trade history |
| **Candles** | OHLCV bars (1m, 5m, 15m, 1h, 4h, 1d) |
| **24h High/Low/Volume** | Rolling 24-hour statistics |

#### HTTP API (port 8082)

```bash
GET /api/v1/market/ticker/BTC/USDT
GET /api/v1/market/tickers
GET /api/v1/market/orderbook/BTC/USDT
GET /api/v1/market/trades/BTC/USDT?limit=50
GET /api/v1/market/candles/BTC/USDT?interval=1m&limit=100
GET /api/v1/market/stats/BTC/USDT
```

#### Run market-data

```bash
KAFKA_BROKERS=localhost:9092 go run ./market-data/cmd/market-data
```

## Ledger Service

Location: `ledger-service/`

Double-entry accounting ledger. Every trade creates a **balanced journal** where debits = credits per asset.

### Alice / Bob example

```
Before:  Alice 1000 USDT  |  Bob 1 BTC
Trade:   Alice buys 0.1 BTC @ 10000 USDT
After:   Alice +0.1 BTC, -1000 USDT  |  Bob -0.1 BTC, +1000 USDT
```

### Journal (trade T-1)

| Leg | User | Asset | Type | Amount |
|-----|------|-------|------|--------|
| 1 | alice | BTC | DEBIT | 10 (0.1) |
| 2 | bob | BTC | CREDIT | 10 |
| 3 | bob | USDT | DEBIT | 1000 |
| 4 | alice | USDT | CREDIT | 1000 |

### HTTP API (port 8083)

```bash
# Deposit funds
curl -X POST http://localhost:8083/api/v1/ledger/deposit \
  -H "Content-Type: application/json" \
  -d '{"user_id":"alice","asset":"USDT","amount":1000}'

# Get balance
curl http://localhost:8083/api/v1/ledger/balance/alice/USDT

# Get journal for a trade
curl http://localhost:8083/api/v1/ledger/journal/T-1

# Settle a trade (called by settlement-service)
curl -X POST http://localhost:8083/api/v1/ledger/settle-trade \
  -H "Content-Type: application/json" \
  -d '{"id":"T-1","symbol":"BTC/USDT","buyer_id":"alice","seller_id":"bob","price":10000,"quantity":10}'
```

### Run ledger-service

```bash
go run ./ledger-service/cmd/ledger-service
```

Trade settlement is handled by settlement-service via `POST /api/v1/ledger/settle-trade`.

### Withdrawal ledger APIs (port 8083)

Used by wallet-service to reserve funds before broadcasting on-chain:

```bash
# Reserve funds for withdrawal
curl -X POST http://localhost:8083/api/v1/ledger/reserve \
  -H "Content-Type: application/json" \
  -d '{"user_id":"alice","asset":"USDT","amount":500,"ref_id":"wd-1"}'

# Release reservation (on broadcast failure)
curl -X POST http://localhost:8083/api/v1/ledger/release \
  -H "Content-Type: application/json" \
  -d '{"user_id":"alice","asset":"USDT","amount":500,"ref_id":"wd-1"}'

# Debit locked funds (after successful broadcast)
curl -X POST http://localhost:8083/api/v1/ledger/debit-locked \
  -H "Content-Type: application/json" \
  -d '{"user_id":"alice","asset":"USDT","amount":500,"ref_id":"wd-1"}'
```

### Run ledger tests

```bash
go test ./ledger-service/tests/ -v
```

## Wallet Service

Location: `wallet-service/`

Handles **blockchain-facing wallet operations** while keeping **ledger balance** as the internal source of truth for user funds.

### Key separation

| Concept | Responsibility |
|---------|----------------|
| **Blockchain Wallet** | On-chain deposit address (per user/asset/chain) |
| **Ledger Balance** | Internal available/locked balance (managed by ledger-service) |

### Deposit flow

```
Blockchain Wallet
       │
       ▼
On-chain Deposit
       │
       ▼
Blockchain Service (monitors chain)
       │
       ▼
wallet-service ConfirmDeposit
       │
       ▼
ledger-service Deposit
       │
       ▼
User Available Balance
```

### Withdrawal flow

```
User
 │
 ▼
Withdrawal Request
 │
 ▼
Risk checks (min/max, balance)
 │
 ▼
Ledger reservation (lock funds)
 │
 ▼
Blockchain Service (broadcast tx)
 │
 ▼
Ledger debit-locked (finalize)
 │
 ▼
Blockchain transaction
```

### HTTP API (port 8084)

```bash
# Create deposit address
curl -X POST http://localhost:8084/api/v1/wallet/address \
  -H "Content-Type: application/json" \
  -d '{"user_id":"alice","asset":"USDT","chain":"ethereum"}'

# Confirm deposit (called by blockchain-service after on-chain confirmation)
curl -X POST http://localhost:8084/api/v1/wallet/deposit/confirm \
  -H "Content-Type: application/json" \
  -d '{"tx_hash":"0xabc","to_address":"<address>","amount":1000,"confirmations":12}'

# Request withdrawal
curl -X POST http://localhost:8084/api/v1/wallet/withdraw \
  -H "Content-Type: application/json" \
  -d '{"user_id":"alice","asset":"USDT","amount":500,"to_address":"0xExternal..."}'

# Get ledger balance (proxied from ledger-service)
curl http://localhost:8084/api/v1/wallet/balance/alice/USDT

# Get withdrawal status
curl http://localhost:8084/api/v1/wallet/withdrawal/<withdrawal_id>
```

### Run wallet-service

```bash
LEDGER_URL=http://localhost:8083 BLOCKCHAIN_URL=http://localhost:8085 go run ./wallet-service/cmd/wallet-service
```

### Run wallet tests

```bash
go test ./wallet-service/tests/ -v
```

## Blockchain Service

Location: `blockchain-service/`

Isolates all EVM/blockchain infrastructure from core exchange services. **order-service** and **matching-engine** never call Ethereum RPC directly — only blockchain-service does.

### Architecture

```
CEX Services (wallet-service, order-service, matching-engine)
     │
     ▼  HTTP only — no direct RPC
Blockchain Service (:8085)
     │
     ├── EVM Provider (abstracted — replaceable)
     ├── Wallet (custody address generation)
     ├── Deposit monitor (watches addresses → wallet-service)
     ├── Withdrawal (broadcast on-chain)
     └── Transaction tracking (pending → confirmed)
             │
             ▼
           EVM
```

### HTTP API (port 8085)

```bash
# Generate custody deposit address
curl -X POST http://localhost:8085/api/v1/blockchain/address \
  -H "Content-Type: application/json" \
  -d '{"user_id":"alice","asset":"USDT","chain":"ethereum"}'

# Validate address
curl -X POST http://localhost:8085/api/v1/blockchain/address/validate \
  -H "Content-Type: application/json" \
  -d '{"address":"0xabc1234567890"}'

# Broadcast withdrawal
curl -X POST http://localhost:8085/api/v1/blockchain/withdraw \
  -H "Content-Type: application/json" \
  -d '{"asset":"USDT","to_address":"0xrecipient...","amount":500}'

# Track transaction
curl http://localhost:8085/api/v1/blockchain/transaction/0xwithdraw-USDT-500-1
```

### Deposit flow (automated)

1. wallet-service requests address from blockchain-service
2. blockchain-service registers address for deposit monitoring
3. Deposit monitor detects on-chain transfer (confirmations met)
4. blockchain-service calls wallet-service `deposit/confirm`
5. wallet-service credits ledger

### Run blockchain-service

```bash
WALLET_SERVICE_URL=http://localhost:8084 go run ./blockchain-service/cmd/blockchain-service
```

Env vars: `EVM_RPC_URL` (future real RPC), `REQUIRED_CONFIRMATIONS` (default 12).

### Run blockchain tests

```bash
go test ./blockchain-service/tests/ -v
```

## Admin Service

Location: `admin-service/`

Operations control plane for the exchange. Provides a unified admin API for users, markets, trading pairs, fees, deposits, withdrawals, risk rules, KYC review, and system health — without coupling to other services' internal packages (HTTP only).

### Dashboard & system status

```bash
# Dashboard stats (users, volumes)
curl -H "X-Admin-API-Key: admin-dev-key" http://localhost:8087/api/v1/admin/dashboard

# Probe matching engine, database, kafka, blockchain RPC
curl -H "X-Admin-API-Key: admin-dev-key" http://localhost:8087/api/v1/admin/system/status
```

Example dashboard response:

```
Users              125,320
Active Users         8,430
BTC Volume         $12.4M
ETH Volume          $7.8M
```

System status maps health probes to:

| Component | Probed service |
|-----------|----------------|
| Matching Engine | order-service |
| Database | ledger-service |
| Kafka | market-data |
| Blockchain RPC | blockchain-service |

### HTTP API (port 8087)

All routes under `/api/v1/admin/` require header `X-Admin-API-Key` (default dev key: `admin-dev-key`). `/health` is public.

| Area | Endpoints |
|------|-----------|
| Dashboard | `GET /dashboard`, `GET /system/status` |
| Users | `GET /users`, `GET /users/:id`, `PATCH /users/:id` |
| Markets | `GET|POST /markets`, `PATCH /fees/:symbol` |
| Deposits | `GET /deposits` |
| Withdrawals | `GET /withdrawals`, `POST /withdrawals/:id/approve\|reject` |
| KYC | `GET /kyc`, `POST /kyc/:id/approve\|reject` |
| Risk | `GET /risk/rules`, `PATCH /risk/rules/:id` |

```bash
# Suspend a user
curl -X PATCH http://localhost:8087/api/v1/admin/users/usr-001 \
  -H "X-Admin-API-Key: admin-dev-key" \
  -H "Content-Type: application/json" \
  -d '{"status":"SUSPENDED"}'

# Update trading pair fees (symbol URL-encoded: BTC%2FUSDT)
curl -X PATCH http://localhost:8087/api/v1/admin/fees/BTC%2FUSDT \
  -H "X-Admin-API-Key: admin-dev-key" \
  -H "Content-Type: application/json" \
  -d '{"maker_fee_bps":5,"taker_fee_bps":10}'

# Approve a pending withdrawal
curl -X POST http://localhost:8087/api/v1/admin/withdrawals/wd-001/approve \
  -H "X-Admin-API-Key: admin-dev-key"
```

Via API gateway:

```bash
curl -H "X-Admin-API-Key: admin-dev-key" http://localhost/api/v1/admin/dashboard
```

### Run admin-service

```bash
ADMIN_API_KEY=admin-dev-key \
ORDER_SERVICE_URL=http://localhost:8081 \
LEDGER_SERVICE_URL=http://localhost:8083 \
BLOCKCHAIN_SERVICE_URL=http://localhost:8085 \
go run ./admin-service/cmd/admin-service
```

### Run admin tests

```bash
go test ./admin-service/tests/ -v
```

## Settlement Service

Location: `settlement-service/`

Finalizes trades after matching. Computes asset transfers, posts to the ledger, and publishes settlement events.

### Flow

```
Matching Engine
       │
       ▼
    Trade (Kafka: trades)
       │
       ▼
Settlement Service
       │
       ├── Buyer gets base asset (minus fee)
       ├── Seller gets quote currency (minus fee)
       ├── Fees routed to exchange account
       └── Ledger updated (via ledger-service HTTP API)
       │
       ▼
Settlement event (Kafka: settlement)
       │
       ▼
Notification Service
```

### Alice / Bob example

```
Trade: Alice buys 0.1 BTC @ 10000 USDT (qty=10, scale=100)

Settlement:
  Alice receives: 10 BTC (base)
  Bob receives:   1000 USDT (quote)
  Ledger journal posted with balanced double-entry legs
```

### Run settlement-service

```bash
# Terminal 1 — ledger HTTP API
go run ./ledger-service/cmd/ledger-service

# Terminal 2 — settlement consumer
KAFKA_BROKERS=localhost:9092 LEDGER_URL=http://localhost:8083 go run ./settlement-service/cmd/settlement-service
```

### Run settlement tests

```bash
go test ./settlement-service/tests/ -v
```

### Run other consumers

```bash
go run ./ledger-service/cmd/ledger-service
KAFKA_BROKERS=localhost:9092 LEDGER_URL=http://localhost:8083 go run ./settlement-service/cmd/settlement-service
KAFKA_BROKERS=localhost:9092 go run ./notification-service/cmd/notification-service
```

### Run order-service with Kafka publishing

```bash
KAFKA_BROKERS=localhost:9092 go run ./order-service/cmd/order-service
```

### Run event tests

```bash
go test ./matching-engine/tests/ ./ledger-service/tests/ ./wallet-service/tests/ ./blockchain-service/tests/ ./api-gateway/tests/ ./auth-service/tests/ ./user-service/tests/ ./pkg/jwt/ ./admin-service/tests/ ./settlement-service/tests/ ./market-data/tests/ ./notification-service/tests/ -v
```

## On-Chain Contracts

Location: `contracts/`

Solidity contracts for **on-chain custody only** — not order matching.

```
Order → Off-chain Matching Engine → Off-chain Ledger → Off-chain Settlement
                                                              │
On-chain (contracts/): Deposits, Withdrawals, Treasury, Tokens
```

| Contract | Purpose |
|----------|---------|
| `IERC20` | Token interface |
| `CEXToken` | ERC-20 assets (mock USDT, WBTC) |
| `CEXVault` | User deposits + operator withdrawals |
| `Treasury` | Fee reserves and cold-storage sweeps |
| `Exchange` | Deploys vault + treasury (NOT an order book) |

See [contracts/README.md](contracts/README.md) for architecture and deployment.

```bash
cd contracts && make install && make test
```

## Project layout

```
matching-engine/     ← order book, engine, settlement, Kafka publisher
api-gateway/         ← front door: routing, CORS, rate limits, auth forwarding (port 8080)
auth-service/        ← register, login, JWT, refresh tokens, sessions (port 8090)
user-service/        ← user profiles in PostgreSQL (port 8091)
frontend/            ← Next.js trading UI (React, 3D hero, API integration)
contracts/           ← Solidity: vault, treasury, tokens (NOT order matching)
order-service/       ← validation pipeline + HTTP API
market-data/         ← Kafka consumer + HTTP API (ticker, book, trades, candles, 24h stats)
ledger-service/      ← double-entry ledger HTTP API (settle-trade, deposit, balances)
wallet-service/      ← blockchain wallets, deposits, withdrawals
blockchain-service/  ← EVM isolation: wallets, deposits, withdrawals, tx tracking (port 8085)
admin-service/       ← operations dashboard: users, markets, fees, KYC, risk, system status (port 8087)
settlement-service/  ← Kafka consumer (trades → ledger → settlement events)
notification-service/← Kafka consumer (orders, trades, settlement)
pkg/contracts/       ← on-chain event signatures for blockchain-service
pkg/events/          ← shared event types + topics
pkg/kafka/           ← Kafka producer/consumer + in-memory bus
pkg/redis/           ← cache, session, rate limit, locks, pub/sub
docker/              ← Docker Compose: all services + postgres, redis, kafka, nginx, monitoring
...
```

## Frontend

Location: `frontend/`

Next.js 14 UI connected to the full backend through **nginx → api-gateway**:

```
Frontend (/api/*) → nginx :80 → api-gateway :8080 → microservices
```

- **Auth**: `/login`, `/register` → auth-service (JWT persisted in browser)
- **Trade**: orders with `Authorization: Bearer` → order-service (BTC/USDT → binance-adapter when `ROUTE_BINANCE=1`)
- **Markets**: market-data tickers
- **Wallet**: ledger balances, deposits, withdrawals

```bash
docker compose -f docker/docker-compose.yml up --build
cd frontend && npm install --legacy-peer-deps && npm run dev
```

See [frontend/README.md](frontend/README.md).

## Docker

Location: `docker/`

Run the full CEX stack as containers:

```bash
docker compose -f docker/docker-compose.yml up --build
```

| URL | Service |
|-----|---------|
| http://localhost | Nginx (frontend + API gateway) |
| http://localhost:3001 | Grafana (admin / admin) |
| http://localhost:3002 | Frontend (Docker) |
| http://localhost:9090 | Prometheus |

Infrastructure only (for local Go dev):

```bash
docker compose -f docker/docker-compose.infra.yml up -d
```

See [docker/README.md](docker/README.md) for full container map and gateway routes.

## Build

```bash
go build ./...
go test ./...
```
