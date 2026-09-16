package health

import (
	"log"
	"net/http"
)

// Serve starts a background HTTP server exposing GET /health.
func Serve(addr string) {
	if addr == "" {
		return
	}
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})
		log.Printf("health endpoint listening on %s", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Printf("health server error: %v", err)
		}
	}()
}

// Register adds GET /health to an existing mux.
func Register(mux *http.ServeMux) {
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}
