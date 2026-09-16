package main

import (
	"log"
	"net/http"
)

// blockchain-service is a placeholder for on-chain monitoring and transaction broadcast.
// In production it watches deposit addresses and calls wallet-service deposit/confirm webhook.
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	log.Println("blockchain-service placeholder listening on :8085")
	log.Fatal(http.ListenAndServe(":8085", mux))
}
