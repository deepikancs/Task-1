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

	"organisation-task-control-centre/backend/internal/config"
	"organisation-task-control-centre/backend/internal/database"
	"organisation-task-control-centre/backend/internal/handlers"
	"organisation-task-control-centre/backend/internal/repository"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[FATAL] Configuration error: %v", err)
	}

	var store repository.Store

	if cfg.DemoMode {
		log.Println("[INFO] DEMO_MODE=true: Starting Go API with in-memory demo data (MySQL connection skipped; changes are temporary).")
		store = repository.NewDemoTaskRepository(cfg.DefaultDemoToday)
	} else {
		db, err := database.Connect(cfg.MySQLDSN())
		if err != nil {
			log.Fatalf(
				"[FATAL] Database connection failed (host=%s:%s db=%s user=%s).\n"+
					"  -> Cause: %v\n"+
					"  -> Fix: Open 'backend/.env' and make sure DB_HOST, DB_PORT, DB_NAME, DB_USER, and DB_PASSWORD match your SQL Studio connection (or set DEMO_MODE=true).",
				cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBUser, err,
			)
		}
		defer func() {
			log.Println("[INFO] Closing MySQL connection pool...")
			_ = db.Close()
		}()

		log.Printf("[INFO] Connected to MySQL database %q on %s:%s", cfg.DBName, cfg.DBHost, cfg.DBPort)
		store = repository.NewTaskRepository(db, cfg.DefaultDemoToday)
	}

	apiHandler := handlers.NewHandler(store, cfg.DefaultDemoToday, cfg.DemoMode)

	mux := http.NewServeMux()
	apiHandler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      withCORSAndLogging(cfg.CORSAllowedOrigin, mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Listen for OS interrupt signals for graceful shutdown
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("[INFO] Organisation Task Control Centre Go API listening on http://localhost:%s", cfg.ServerPort)
		log.Printf("[INFO] Allowed CORS Origin: %s", cfg.CORSAllowedOrigin)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[FATAL] HTTP server error: %v", err)
		}
	}()

	sig := <-shutdownChan
	log.Printf("[INFO] Received signal %v, initiating graceful shutdown...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("[ERROR] Graceful shutdown failed: %v", err)
	} else {
		log.Println("[INFO] HTTP server stopped cleanly.")
	}
}

func withCORSAndLogging(allowedOrigin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
		log.Printf("[HTTP] %s %s (%v)", r.Method, r.URL.RequestURI(), time.Since(start))
	})
}
