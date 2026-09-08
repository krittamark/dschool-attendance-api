package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dschool-attendance-api/internal/api"
	"dschool-attendance-api/internal/auth"
	"dschool-attendance-api/internal/cache"
	"dschool-attendance-api/internal/client"
	"dschool-attendance-api/internal/config"
)

func main() {
	log.Println("=== Starting dschool Attendance API Server ===")

	cfg := config.LoadConfig()

	// Initialize Dynamic Key Store
	log.Println("Initializing API Key Store...")
	keyStore, err := auth.NewKeyStore(cfg.KeysFilePath, cfg.ApiKey, cfg.AdminApiKey)
	if err != nil {
		log.Fatalf("Failed to initialize API key store: %v", err)
	}
	log.Printf("API Key Store ready! Total keys loaded: %d\n", len(keyStore.ListKeys()))

	// Initialize In-Memory Cache
	memCache := cache.NewMemoryCache()
	log.Println("In-Memory Cache initialized!")

	log.Println("Connecting to dschool and logging in...")
	dsClient, err := client.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize dschool client: %v", err)
	}
	log.Println("Successfully authenticated with dschool session!")

	handler := api.NewHandler(dsClient, cfg, keyStore, memCache)
	router := api.SetupRouter(handler)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown channel
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server listening on http://localhost:%s\n", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}
	log.Println("Server gracefully stopped.")
}
