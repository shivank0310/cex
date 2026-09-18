# CEX Frontend

Next.js 14 trading platform UI for the CEX microservices backend.

Inspired by professional CEX platforms ([Codono CEX Software](https://codono.com/solutions/cex-exchange-software)) with off-chain matching and on-chain custody architecture.

## Stack

- **Next.js 14** (App Router)
- **TypeScript** + **Tailwind CSS**
- **React Three Fiber** — 3D hero scene
- **Framer Motion** — animations
- **Zustand** — trading state

## Structure

```
src/
├── app/              # Pages (landing, trade, markets, wallet)
├── components/
│   ├── landing/      # Hero, features, architecture
│   ├── trade/        # Order book, chart, order form
│   ├── wallet/       # Deposit, withdraw, balances
│   ├── three/        # 3D WebGL scene
│   ├── layout/       # Header, footer
│   └── ui/           # Shared UI primitives
├── lib/api/          # Backend API clients
├── hooks/            # useMarketData polling
├── store/            # Zustand trading store
└── types/            # TypeScript interfaces
```

## Pages

| Route | Description |
|-------|-------------|
| `/` | Landing — 3D hero, features, architecture |
| `/trade` | Trading terminal — chart, order book, place orders |
| `/markets` | Market overview — tickers table |
| `/wallet` | Balances, deposit address, withdrawals |

## Development

```bash
cd frontend
cp .env.local.example .env.local

# Start backend (from repo root)
docker compose -f docker/docker-compose.infra.yml up -d
# + order-service, market-data, ledger, wallet, blockchain

npm install --legacy-peer-deps
npm run dev
```

Open http://localhost:3000

API requests go to `/api/*` and are proxied to `API_PROXY_URL` (default `http://localhost` nginx gateway).

## Demo auth tokens

| User | Token |
|------|-------|
| user-a | `token-user-a` |
| seller-1 | `token-seller` |

Orders use `Authorization` header automatically from the trading store.

## Docker

```bash
docker build -f frontend/Dockerfile -t cex/frontend ./frontend
docker run -p 3000:3000 -e API_PROXY_URL=http://host.docker.internal:80 cex/frontend
```

## Backend integration

| Feature | API |
|---------|-----|
| Place order | `POST /api/v1/orders` |
| Market ticker | `GET /api/v1/market/ticker/BTC/USDT` |
| Order book | `GET /api/v1/market/orderbook/BTC/USDT` |
| Wallet balance | `GET /api/v1/wallet/balance/{user}/{asset}` |
| Deposit address | `POST /api/v1/wallet/address` |
