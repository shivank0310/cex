package evm

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/shivank0310/cex.git/blockchain-service/internal/model"
)

// MockProvider simulates an EVM chain for development and tests.
type MockProvider struct {
	mu          sync.RWMutex
	chain       string
	blockNum    uint64
	addrSeq     uint64
	txSeq       uint64
	receipts    map[string]*Receipt
	deposits    []model.DepositEvent
	processed   map[string]bool
}

func NewMockProvider(chain string) *MockProvider {
	return &MockProvider{
		chain:     chain,
		blockNum:  1000,
		receipts:  make(map[string]*Receipt),
		processed: make(map[string]bool),
	}
}

func (m *MockProvider) Chain() string { return m.chain }

func (m *MockProvider) ValidateAddress(address string) bool {
	return strings.HasPrefix(address, "0x") && len(address) >= 10
}

func (m *MockProvider) GenerateAddress(_ context.Context, userID, asset string) (string, error) {
	id := atomic.AddUint64(&m.addrSeq, 1)
	return fmt.Sprintf("0x%s%s%04d", shorten(userID, 4), shorten(asset, 4), id), nil
}

func (m *MockProvider) BroadcastWithdrawal(_ context.Context, asset, toAddress string, amount int64) (string, error) {
	if !m.ValidateAddress(toAddress) {
		return "", fmt.Errorf("invalid address: %s", toAddress)
	}
	if amount <= 0 {
		return "", fmt.Errorf("amount must be positive")
	}

	id := atomic.AddUint64(&m.txSeq, 1)
	txHash := fmt.Sprintf("0xwithdraw-%s-%d-%d", asset, amount, id)

	m.mu.Lock()
	m.blockNum++
	m.receipts[txHash] = &Receipt{
		TxHash:        txHash,
		BlockNumber:   m.blockNum,
		Confirmations: 1,
		Status:        model.TxConfirmed,
	}
	m.mu.Unlock()
	return txHash, nil
}

func (m *MockProvider) GetReceipt(_ context.Context, txHash string) (*Receipt, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.receipts[txHash]
	if !ok {
		return nil, fmt.Errorf("transaction not found: %s", txHash)
	}
	confirmations := int(m.blockNum - r.BlockNumber + 1)
	if confirmations < 1 {
		confirmations = 1
	}
	return &Receipt{
		TxHash:        r.TxHash,
		BlockNumber:   r.BlockNumber,
		Confirmations: confirmations,
		Status:        r.Status,
	}, nil
}

func (m *MockProvider) GetBlockNumber(_ context.Context) (uint64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.blockNum, nil
}

func (m *MockProvider) ScanDeposits(_ context.Context, addresses []string, fromBlock uint64) ([]model.DepositEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	watched := make(map[string]bool, len(addresses))
	for _, a := range addresses {
		watched[a] = true
	}

	var out []model.DepositEvent
	remaining := make([]model.DepositEvent, 0, len(m.deposits))
	for _, dep := range m.deposits {
		if !watched[dep.ToAddress] {
			remaining = append(remaining, dep)
			continue
		}
		if m.processed[dep.TxHash] {
			continue
		}
		if dep.BlockNumber < fromBlock {
			continue
		}
		dep.Confirmations = int(m.blockNum - dep.BlockNumber + 1)
		out = append(out, dep)
	}
	m.deposits = remaining
	return out, nil
}

// SimulateDeposit injects an on-chain deposit for tests and local development.
func (m *MockProvider) SimulateDeposit(toAddress, asset string, amount int64) model.DepositEvent {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.blockNum++
	id := atomic.AddUint64(&m.txSeq, 1)
	dep := model.DepositEvent{
		TxHash:        fmt.Sprintf("0xdeposit-%s-%d", asset, id),
		ToAddress:     toAddress,
		Asset:         asset,
		Amount:        amount,
		BlockNumber:   m.blockNum,
		Confirmations: 1,
	}
	m.deposits = append(m.deposits, dep)
	m.receipts[dep.TxHash] = &Receipt{
		TxHash:        dep.TxHash,
		BlockNumber:   dep.BlockNumber,
		Confirmations: 1,
		Status:        model.TxConfirmed,
	}
	return dep
}

func (m *MockProvider) MarkProcessed(txHash string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processed[txHash] = true
}

func (m *MockProvider) AdvanceBlocks(n uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.blockNum += n
}

func shorten(s string, n int) string {
	s = strings.ToLower(strings.ReplaceAll(s, " ", ""))
	if len(s) >= n {
		return s[:n]
	}
	return fmt.Sprintf("%s%0*s", s, n-len(s), "")
}

// NewProvider returns the configured EVM provider. RPC URL wiring can be added later.
func NewProvider(chain, rpcURL string) Provider {
	if rpcURL != "" {
		// Future: return HTTP JSON-RPC provider. Mock keeps blockchain replaceable in dev.
	}
	return NewMockProvider(chain)
}
