package testserver

import (
	"net/http"
	"net/http/httptest"

	"github.com/shivank0310/cex.git/ledger-service/internal/engine"
	"github.com/shivank0310/cex.git/ledger-service/internal/handler"
	"github.com/shivank0310/cex.git/ledger-service/internal/repository"
	"github.com/shivank0310/cex.git/ledger-service/internal/service"
)

// LedgerService exposes ledger operations for integration tests.
type LedgerService = service.LedgerService

// New starts an in-memory ledger HTTP server for integration tests.
func New() (*httptest.Server, *LedgerService) {
	repo := repository.NewLedgerRepository()
	eng := engine.NewDoubleEntryEngine(repo)
	svc := service.NewLedgerService(repo, eng, nil)
	h := handler.NewLedgerHandler(svc)

	mux := http.NewServeMux()
	h.Register(mux)
	return httptest.NewServer(mux), svc
}
