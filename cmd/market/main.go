package main

import (
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dvtkrlbs/market/internal/api"
	"github.com/dvtkrlbs/market/internal/database"
	"github.com/dvtkrlbs/market/internal/services/auth"
	authservice "github.com/dvtkrlbs/market/internal/services/auth"
	productsservice "github.com/dvtkrlbs/market/internal/services/products"
	"github.com/dvtkrlbs/market/internal/services/stores"
	storesservice "github.com/dvtkrlbs/market/internal/services/stores"
)

const defaultDatabaseURL = "postgres://market:market@localhost:5432/market?sslmode=disable"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return usage()
	}
	if args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		return usage()
	}
	if err := loadDevelopmentEnv(); err != nil {
		return err
	}
	databaseURL := envOr("DATABASE_URL", defaultDatabaseURL)
	if args[0] == "migrate" {
		if len(args) != 1 {
			return fmt.Errorf("migrate does not accept arguments")
		}
		return database.Migrate(databaseURL)
	}
	client, err := database.Open(databaseURL)
	if err != nil {
		return err
	}
	defer client.Close()

	switch args[0] {
	case "serve":
		return serve(authservice.New(client), productsservice.New(client), stores.New(client), args[1:])
	default:
		return fmt.Errorf("unknown command %q\n\n%w", args[0], usage())
	}
}

func serve(authentication *authservice.Service, productService *productsservice.Service, storeService *storesservice.Service, args []string) error {
	flags := flag.NewFlagSet("serve", flag.ContinueOnError)
	addr := flags.String("addr", envOr("HTTP_ADDR", ":8080"), "HTTP listen address")
	if err := flags.Parse(args); err != nil {
		return err
	}

	key, err := base64.StdEncoding.Strict().DecodeString(os.Getenv("SESSION_SIGNING_KEY"))
	if err != nil {
		return fmt.Errorf("SESSION_SIGNING_KEY must be a base64-encoded key: %w", err)
	}
	signer, err := auth.NewCookieSigner(key)
	if err != nil {
		return fmt.Errorf("SESSION_SIGNING_KEY: %w", err)
	}
	handler, err := api.NewRouter(authentication, productService, signer, storeService)
	if err != nil {
		return err
	}
	handler, err = api.WithCORS(handler, strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ","))
	if err != nil {
		return err
	}
	server := &http.Server{Addr: *addr, Handler: api.WithRequestLogging(handler), ReadHeaderTimeout: 5 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	slog.Info("API listening", "address", *addr)
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func usage() error {
	message := strings.TrimSpace(`market - small bookkeeping application

Usage:
  market migrate
  market serve [--addr :8080]

Environment:
  ENVIRONMENT   Runtime environment (default DEVELOPMENT); only DEVELOPMENT loads .env
  DATABASE_URL  PostgreSQL connection string
  HTTP_ADDR     Default server listen address
  SESSION_SIGNING_KEY  Base64-encoded session cookie signing key (at least 32 bytes)
  CORS_ALLOWED_ORIGINS  Comma-separated frontend origins allowed to send cookies`)
	fmt.Fprintln(os.Stderr, message)
	return nil
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
