// Command chater is the entrypoint for the chater messenger backend.
//
// Step 1 (base) wires only configuration, a stdlib HTTP server, structured
// logging and graceful shutdown. Rooms, messages, persistence and websockets
// arrive in later steps.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/omnifield/chater/internal/config"
	"github.com/omnifield/chater/internal/httpapi"
	"github.com/omnifield/chater/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(context.Background(), logger); err != nil {
		logger.Error("chater exited with error", "err", err)
		os.Exit(1)
	}
}

// run owns the server lifecycle: load env-only config, serve HTTP, and shut
// down gracefully on SIGINT/SIGTERM. It returns an error rather than exiting so
// it stays testable.
func run(ctx context.Context, logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Open the database and apply migrations on startup (idempotent goose up).
	// Rationale: dev-first, one moving part — the service is always schema-ready
	// without an out-of-band migrate step. A standalone `migrate` command can be
	// added later if prod ops want migrations gated separately from rollout.
	db, err := store.Open(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer func() {
		if cerr := db.Close(); cerr != nil {
			logger.Warn("closing database", "err", cerr)
		}
	}()
	if err := store.Migrate(ctx, db); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	logger.Info("migrations applied", "db", cfg.DBPath)

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           httpapi.NewRouter(logger),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("chater listening", "addr", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
		logger.Info("shutdown signal received, draining")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	logger.Info("chater stopped cleanly")
	return nil
}
