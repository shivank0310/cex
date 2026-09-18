package main

import (
	"log"
	"net/http"
	"os"

	"github.com/shivank0310/cex.git/user-service/internal/config"
	"github.com/shivank0310/cex.git/user-service/internal/database"
	"github.com/shivank0310/cex.git/user-service/internal/handler"
	"github.com/shivank0310/cex.git/user-service/internal/repository"
	"github.com/shivank0310/cex.git/user-service/internal/service"
)

func main() {
	cfg := config.Default()
	if addr := os.Getenv("HTTP_ADDR"); addr != "" {
		cfg.HTTPAddr = addr
	}
	if url := os.Getenv("DATABASE_URL"); url != "" {
		cfg.DatabaseURL = url
	}

	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connect failed: %v", err)
	}
	defer db.Close()

	repo := repository.NewUserRepository(db)
	svc := service.NewUserService(repo)
	userHandler := handler.NewUserHandler(svc)

	mux := http.NewServeMux()
	userHandler.Register(mux)

	log.Printf("user-service listening on %s", cfg.HTTPAddr)
	log.Fatal(http.ListenAndServe(cfg.HTTPAddr, mux))
}
