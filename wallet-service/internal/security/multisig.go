package security

import (
	"fmt"
	"sync"
	"time"
)

// MultisigKey represents a signing key in the multisig wallet.
type MultisigKey struct {
	ID       string
	Type     string // "HSM", "HSM", "COLD"
	Label    string
}

// MultisigSignature records a co-signer approval.
type MultisigSignature struct {
	KeyID        string
	KeyType      string
	SignerID     string
	Signature    string
	SignedAt     time.Time
}

// MultisigWallet manages N-of-M signing for large treasury withdrawals.
//
// Example layout:
//   Key A → HSM (hot operations)
//   Key B → HSM (warm operations)
//   Key C → Cold/independent custody
type MultisigWallet struct {
	mu              sync.RWMutex
	keys            []MultisigKey
	requiredSigs    int
	pending         map[string][]MultisigSignature // withdrawalID -> signatures
}

func NewMultisigWallet() *MultisigWallet {
	return &MultisigWallet{
		keys: []MultisigKey{
			{ID: "key-a", Type: "HSM", Label: "HSM Key A (operations)"},
			{ID: "key-b", Type: "HSM", Label: "HSM Key B (compliance)"},
			{ID: "key-c", Type: "COLD", Label: "Cold Storage Key C"},
		},
		requiredSigs: 2, // 2-of-3
		pending:      make(map[string][]MultisigSignature),
	}
}

func (m *MultisigWallet) RequiredSignatures() int {
	return m.requiredSigs
}

func (m *MultisigWallet) Keys() []MultisigKey {
	return append([]MultisigKey(nil), m.keys...)
}

// Sign adds a co-signer signature for a withdrawal.
func (m *MultisigWallet) Sign(withdrawalID, keyID, signerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var key *MultisigKey
	for _, k := range m.keys {
		if k.ID == keyID {
			key = &k
			break
		}
	}
	if key == nil {
		return fmt.Errorf("unknown multisig key: %s", keyID)
	}

	for _, sig := range m.pending[withdrawalID] {
		if sig.KeyID == keyID {
			return fmt.Errorf("key %s already signed", keyID)
		}
	}

	m.pending[withdrawalID] = append(m.pending[withdrawalID], MultisigSignature{
		KeyID: keyID, KeyType: key.Type, SignerID: signerID,
		Signature: fmt.Sprintf("multisig-%s-%s", keyID, withdrawalID),
		SignedAt:  time.Now().UTC(),
	})
	return nil
}

// IsFullySigned returns true when enough co-signatures are collected.
func (m *MultisigWallet) IsFullySigned(withdrawalID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.pending[withdrawalID]) >= m.requiredSigs
}

func (m *MultisigWallet) GetSignatures(withdrawalID string) []MultisigSignature {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]MultisigSignature(nil), m.pending[withdrawalID]...)
}

// SignAndBroadcast performs final HSM signing after multisig quorum is met.
func (m *MultisigWallet) SignAndBroadcast(withdrawalID, asset, toAddress string, amount int64) (SignedTransaction, error) {
	if !m.IsFullySigned(withdrawalID) {
		sigs := m.GetSignatures(withdrawalID)
		return SignedTransaction{}, fmt.Errorf("multisig quorum not met: %d/%d signatures", len(sigs), m.requiredSigs)
	}

	// Final signing uses cold-storage key after quorum.
	coldSigner := NewMockHSMSigner("hsm-key-c-cold")
	return coldSigner.Sign(asset, toAddress, amount, withdrawalID)
}
