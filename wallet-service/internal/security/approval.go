package security

import (
	"fmt"
	"sync"
	"time"
)

// ApprovalRecord tracks manual approval for a withdrawal.
type ApprovalRecord struct {
	WithdrawalID string
	ApproverID   string
	Approved     bool
	Note         string
	Timestamp    time.Time
}

// ApprovalStore manages manual approval decisions.
type ApprovalStore struct {
	mu        sync.RWMutex
	approvals map[string][]ApprovalRecord // withdrawalID -> approvals
}

func NewApprovalStore() *ApprovalStore {
	return &ApprovalStore{approvals: make(map[string][]ApprovalRecord)}
}

func (s *ApprovalStore) Approve(withdrawalID, approverID, note string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.approvals[withdrawalID] = append(s.approvals[withdrawalID], ApprovalRecord{
		WithdrawalID: withdrawalID,
		ApproverID:   approverID,
		Approved:     true,
		Note:         note,
		Timestamp:    time.Now().UTC(),
	})
}

func (s *ApprovalStore) Reject(withdrawalID, approverID, note string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.approvals[withdrawalID] = append(s.approvals[withdrawalID], ApprovalRecord{
		WithdrawalID: withdrawalID,
		ApproverID:   approverID,
		Approved:     false,
		Note:         note,
		Timestamp:    time.Now().UTC(),
	})
}

func (s *ApprovalStore) IsApproved(withdrawalID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.approvals[withdrawalID] {
		if a.Approved {
			return true
		}
	}
	return false
}

func (s *ApprovalStore) IsRejected(withdrawalID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.approvals[withdrawalID] {
		if !a.Approved {
			return true
		}
	}
	return false
}

func (s *ApprovalStore) GetApprovals(withdrawalID string) []ApprovalRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]ApprovalRecord(nil), s.approvals[withdrawalID]...)
}

// ApprovalPolicy validates that required approvals are in place.
type ApprovalPolicy struct {
	store *ApprovalStore
}

func NewApprovalPolicy(store *ApprovalStore) *ApprovalPolicy {
	return &ApprovalPolicy{store: store}
}

func (p *ApprovalPolicy) Approve(withdrawalID, approverID, note string) {
	p.store.Approve(withdrawalID, approverID, note)
}

func (p *ApprovalPolicy) Reject(withdrawalID, approverID, note string) {
	p.store.Reject(withdrawalID, approverID, note)
}

func (p *ApprovalPolicy) RequireApproval(withdrawalID string) error {
	if p.store.IsRejected(withdrawalID) {
		return fmt.Errorf("withdrawal rejected by approver")
	}
	if !p.store.IsApproved(withdrawalID) {
		return fmt.Errorf("manual approval required")
	}
	return nil
}
