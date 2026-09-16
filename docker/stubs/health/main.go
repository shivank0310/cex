package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func main() {
	name := os.Getenv("SERVICE_NAME")
	if name == "" {
		name = "cex-stub"
	}
	port := os.Getenv("HTTP_ADDR")
	if port == "" {
		port = ":8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"service": name,
			"status":  "placeholder",
			"message": "service stub — replace with full implementation",
		})
	})

	log.Printf("%s stub listening on %s", name, port)
	log.Fatal(http.ListenAndServe(port, mux))
}
