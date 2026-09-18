# CEX Frontend

Next.js 14 trading platform UI connected to the full CEX backend via **nginx → api-gateway**.

## Stack

- **Next.js 14** (App Router)
- **TypeScript** + **Tailwind CSS**
- **Zustand** — auth + trading state (persisted session)

## Pages

| Route | Description |
|-------|-------------|
| `/` | Landing |
| `/login` | Sign in (auth-service JWT) |
| `/register` | Create account (auth → user-service) |
| `/trade` | Trading terminal (requires auth) |
| `/markets` | Market tickers |
| `/wallet` | Balances, deposit, withdraw (requires auth) |

## Development

```bash
# Full backend stack (from repo root)
docker compose -f docker/docker-compose.yml up --build

cd frontend
npm install --legacy-peer-deps
npm run dev
```

Open http://localhost:3000

API requests use `/api/*` rewrites → `API_PROXY_URL` (default `http://localhost`).

## Auth flow

1. Register or login → `POST /api/v1/auth/register|login`
2. JWT stored in `localStorage` via Zustand persist
3. Orders send `Authorization: Bearer <access_token>` to order-service
4. Wallet uses authenticated user ID from JWT profile

## Binance routing

When `ROUTE_BINANCE=1` on order-service (enabled in Docker Compose), **BTC/USDT** orders route to `binance-adapter-service`. The trade UI shows a **Binance via adapter** badge for that pair.

## Docker

```bash
docker build -f frontend/Dockerfile \
  --build-arg API_PROXY_URL=http://nginx:80 \
  -t cex/frontend ./frontend
```

## Backend integration

| Feature | API |
|---------|-----|
| Register / Login | `/api/v1/auth/*` |
| Place order | `POST /api/v1/orders` |
| Market data | `/api/v1/market/*` |
| Wallet | `/api/v1/wallet/*` |
