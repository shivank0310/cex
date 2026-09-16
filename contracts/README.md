# CEX On-Chain Contracts

Solidity contracts for **on-chain custody** — deposits, withdrawals, treasury, and token operations.

## Important: Off-Chain vs On-Chain

A centralized exchange does **not** need a smart contract for every order.

```
Order
 │
 ▼
Off-chain Matching Engine     (matching-engine/)
 │
 ▼
Off-chain Ledger              (ledger-service/)
 │
 ▼
Off-chain Settlement          (settlement-service/)
```

Blockchain is used for:

| On-chain | Off-chain |
|----------|-----------|
| Deposits (`CEXVault.deposit`) | Order placement |
| Withdrawals (`CEXVault.withdraw`) | Order matching |
| Treasury fee sweeps | Trade settlement |
| Token minting (testnet) | Balance accounting |

`blockchain-service/` is separate from `matching-engine/` for this reason.

## Layout

```
contracts/
├── src/
│   ├── interfaces/IERC20.sol      ← ERC-20 interface
│   ├── token/CEXToken.sol         ← Asset tokens (mock USDT, WBTC)
│   ├── treasury/Treasury.sol      ← Fee reserves
│   └── exchange/
│       ├── CEXVault.sol           ← Deposit/withdraw custody
│       └── Exchange.sol           ← Deploys vault + treasury
├── test/                          ← Foundry tests
└── script/Deploy.s.sol            ← Deployment script
```

## Contracts

### CEXToken

Standard ERC-20 for on-chain assets. Minter role for testnet bootstrap.

### CEXVault

User custody contract:

- `deposit(token, amount)` — user sends tokens, emits `Deposit` event
- `withdraw(to, token, amount, withdrawalId)` — operator only, after off-chain ledger reservation
- `sweepFees(token, amount)` — move fees to treasury

### Treasury

Holds swept fees. Operator can withdraw to cold storage.

### Exchange

Deploys `Treasury` + `CEXVault`. **Does not match orders.**

## Integration with Go Services

```
User deposits USDT on-chain
       │
       ▼
CEXVault.deposit()  →  Deposit event
       │
       ▼
blockchain-service (monitors events)
       │
       ▼
wallet-service ConfirmDeposit
       │
       ▼
ledger-service (off-chain balance)
```

Withdrawal:

```
User requests withdrawal (wallet-service)
       │
       ▼
ledger-service Reserve()
       │
       ▼
blockchain-service → CEXVault.withdraw() (operator)
       │
       ▼
ledger-service DebitLocked()
```

## Setup

```bash
# Install Foundry: https://book.getfoundry.sh/getting-started/installation
cd contracts
forge install foundry-rs/forge-std
forge test
```

## Deploy (testnet)

```bash
export CEX_OPERATOR=0xYourBlockchainServiceHotWallet
forge script script/Deploy.s.sol --rpc-url $EVM_RPC_URL --broadcast
```

Set deployed addresses in blockchain-service:

```bash
VAULT_CONTRACT_ADDRESS=0x...
EVM_RPC_URL=https://...
```
