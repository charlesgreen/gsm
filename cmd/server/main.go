// Package main provides the HTTP server executable for the Google Secret Manager emulator.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charlesgreen/gsm/internal/api/routes"
	"github.com/charlesgreen/gsm/internal/storage"
)

func main() {
	port := getEnvOrDefault("GSM_PORT", "8085")
	host := getEnvOrDefault("GSM_HOST", "0.0.0.0")
	storageFile := os.Getenv("GSM_STORAGE_FILE")
	logLevel := getEnvOrDefault("GSM_LOG_LEVEL", "info")

	fmt.Printf("Starting Google Secret Manager Emulator\n")
	fmt.Printf("Port: %s\n", port)
	fmt.Printf("Host: %s\n", host)
	fmt.Printf("Log Level: %s\n", logLevel)
	if storageFile != "" {
		fmt.Printf("Storage File: %s\n", storageFile)
	}

	var store storage.Storage
	if storageFile != "" {
		persistentStore, err := storage.NewPersistentStorage(storageFile)
		if err != nil {
			log.Fatalf("Failed to create persistent storage: %v", err)
		}
		store = persistentStore
		
		if err := persistentStore.Load(); err != nil {
			log.Printf("Warning: Failed to load existing storage: %v", err)
		}
	} else {
		store = storage.NewMemoryStorage()
	}

	router := routes.SetupRoutes(store)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", host, port),
		Handler: router,
	}

	go func() {
		fmt.Printf("Server starting on http://%s:%s\n", host, port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	if err := store.Close(); err != nil {
		log.Printf("Failed to close storage: %v", err)
	}

	fmt.Println("Server gracefully stopped")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

