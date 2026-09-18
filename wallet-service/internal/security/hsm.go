package security

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
)

// SignedTransaction is the output of HSM signing.
type SignedTransaction struct {
	TxHash    string
	Signature string
	KeyID     string
	SignedAt  string
}

// HSMSigner signs blockchain transactions using hardware security module keys.
// Private keys never leave the HSM boundary.
type HSMSigner interface {
	Sign(asset, toAddress string, amount int64, withdrawalID string) (SignedTransaction, error)
	KeyID() string
}

// MockHSMSigner simulates HSM signing for development.
type MockHSMSigner struct {
	mu    sync.Mutex
	keyID string
}

func NewMockHSMSigner(keyID string) *MockHSMSigner {
	return &MockHSMSigner{keyID: keyID}
}

func (h *MockHSMSigner) KeyID() string { return h.keyID }

func (h *MockHSMSigner) Sign(asset, toAddress string, amount int64, withdrawalID string) (SignedTransaction, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	payload := fmt.Sprintf("%s:%s:%d:%s:%s", asset, toAddress, amount, withdrawalID, h.keyID)
	hash := sha256.Sum256([]byte(payload))
	sig := hex.EncodeToString(hash[:16])

	return SignedTransaction{
		TxHash:    "0xhsm-" + hex.EncodeToString(hash[:8]),
		Signature: sig,
		KeyID:     h.keyID,
		SignedAt:  "mock-hsm",
	}, nil
}

// HSMSignerPool routes signing to the appropriate HSM key based on amount tier.
type HSMSignerPool struct {
	hotSigner  HSMSigner // Key A — hot wallet (small amounts)
	warmSigner HSMSigner // Key B — warm wallet (medium amounts)
}

func NewHSMSignerPool() *HSMSignerPool {
	return &HSMSignerPool{
		hotSigner:  NewMockHSMSigner("hsm-key-a-hot"),
		warmSigner: NewMockHSMSigner("hsm-key-b-warm"),
	}
}

func (p *HSMSignerPool) SignForAmount(asset, toAddress string, amount int64, withdrawalID string, tier string) (SignedTransaction, error) {
	switch tier {
	case "MULTISIG":
		return SignedTransaction{}, fmt.Errorf("multisig withdrawals must use MultisigWallet.Sign")
	default:
		if amount > 100_000 {
			return p.warmSigner.Sign(asset, toAddress, amount, withdrawalID)
		}
		return p.hotSigner.Sign(asset, toAddress, amount, withdrawalID)
	}
}
