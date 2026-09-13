package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"neraca/internal/api"
	"neraca/internal/config"
	"neraca/internal/database"
	"neraca/web"
)

func main() {
	cfg := config.Load()

	// Inisialisasi Database SQLite
	db, err := database.Connect(cfg.DBPath)
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer db.Close()

	// Inisialisasi Static Files (Embedded Vue 3 Dist)
	staticFS, err := web.GetFS()
	if err != nil {
		log.Printf("Warning: failed to load embedded frontend: %v", err)
	}

	// Inisialisasi Router HTTP
	router := api.NewRouter(cfg, db, staticFS)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Jalankan server di goroutine
	go func() {
		log.Printf("✨ Neraca server is running on http://localhost:%s (ENV=%s)", cfg.Port, cfg.Env)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Neraca server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Neraca server stopped gracefully.")
}
