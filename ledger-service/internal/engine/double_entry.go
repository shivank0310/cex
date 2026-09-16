package engine

import (
	"fmt"

	"github.com/shivank0310/cex.git/ledger-service/internal/model"
	"github.com/shivank0310/cex.git/ledger-service/internal/repository"
)

// DoubleEntryEngine posts balanced journals to the ledger repository.
type DoubleEntryEngine struct {
	repo *repository.LedgerRepository
}

func NewDoubleEntryEngine(repo *repository.LedgerRepository) *DoubleEntryEngine {
	return &DoubleEntryEngine{repo: repo}
}

// PostJournal validates and applies a journal atomically.
func (e *DoubleEntryEngine) PostJournal(journal model.Journal) error {
	if err := journal.Validate(); err != nil {
		return err
	}

	if e.repo.IsReferenceProcessed(journal.Reference) && journal.Type == model.JournalTrade {
		return fmt.Errorf("trade already processed: %s", journal.Reference)
	}

	return e.repo.ApplyJournal(journal)
}

// Deposit credits a user account via double-entry (treasury → user).
func (e *DoubleEntryEngine) Deposit(userID, asset string, amount int64, ref string) error {
	if amount <= 0 {
		return fmt.Errorf("deposit amount must be positive")
	}
	journalID := e.repo.NextJournalID()
	journal := BuildDepositJournal(journalID, userID, asset, amount, ref)
	return e.PostJournal(journal)
}
