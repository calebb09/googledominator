package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"googledominator-backend/config"
	"googledominator-backend/db"
	"googledominator-backend/routes"
)

func main() {
	log.Println("==================================================")
	log.Println("     Starting GoogleDominator Backend API         ")
	log.Println("==================================================")

	cfg := config.LoadConfig()

	// Initialize Prisma PostgreSQL database
	databaseClient, err := db.InitDB(cfg)
	if err != nil {
		log.Fatalf("[FATAL] Database initialization failed: %v", err)
	}
	defer databaseClient.Disconnect()

	// Initialize Gin Router
	router := routes.SetupRouter(cfg)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Run server in background goroutine
	go func() {
		log.Printf("[INFO] Server listening on http://localhost:%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[FATAL] Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown listener
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[INFO] Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("[ERROR] Server forced to shutdown: %v", err)
	}

	log.Println("[INFO] Server exited successfully")
}
