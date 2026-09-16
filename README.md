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
go test ./matching-engine/tests/ ./ledger-service/tests/ ./wallet-service/tests/ ./blockchain-service/tests/ ./settlement-service/tests/ ./market-data/tests/ ./notification-service/tests/ -v
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
contracts/           ← Solidity: vault, treasury, tokens (NOT order matching)
order-service/       ← validation pipeline + HTTP API
market-data/         ← Kafka consumer + HTTP API (ticker, book, trades, candles, 24h stats)
ledger-service/      ← double-entry ledger HTTP API (settle-trade, deposit, balances)
wallet-service/      ← blockchain wallets, deposits, withdrawals
blockchain-service/  ← EVM isolation: wallets, deposits, withdrawals, tx tracking (port 8085)
settlement-service/  ← Kafka consumer (trades → ledger → settlement events)
notification-service/← Kafka consumer (orders, trades, settlement)
pkg/contracts/       ← on-chain event signatures for blockchain-service
pkg/events/          ← shared event types + topics
pkg/kafka/           ← Kafka producer/consumer + in-memory bus
pkg/redis/           ← cache, session, rate limit, locks, pub/sub
docker/              ← Docker Compose: all services + postgres, redis, kafka, nginx, monitoring
...
```

## Docker

Location: `docker/`

Run the full CEX stack as containers:

```bash
docker compose -f docker/docker-compose.yml up --build
```

| URL | Service |
|-----|---------|
| http://localhost | Nginx (frontend + API gateway) |
| http://localhost:3000 | Grafana (admin / admin) |
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
