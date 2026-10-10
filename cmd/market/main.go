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
	"text/tabwriter"
	"time"

	"github.com/dvtkrlbs/market/internal/api"
	"github.com/dvtkrlbs/market/internal/auth"
	"github.com/dvtkrlbs/market/internal/bookkeeping"
	"github.com/dvtkrlbs/market/internal/database"
	"github.com/dvtkrlbs/market/internal/money"
	"github.com/dvtkrlbs/market/internal/stores"
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
	service := bookkeeping.New(client)
	authentication := auth.New(client)

	switch args[0] {
	case "serve":
		return serve(service, authentication, stores.New(client), args[1:])
	case "add":
		return add(service, args[1:])
	case "list":
		return list(service, args[1:])
	default:
		return fmt.Errorf("unknown command %q\n\n%w", args[0], usage())
	}
}

func serve(service *bookkeeping.Service, authentication *auth.Service, storeService *stores.Service, args []string) error {
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
	handler, err := api.NewRouter(service, authentication, signer, storeService)
	if err != nil {
		return err
	}
	handler, err = api.WithCORS(handler, strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ","))
	if err != nil {
		return err
	}
	server := &http.Server{Addr: *addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
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

func add(service *bookkeeping.Service, args []string) error {
	flags := flag.NewFlagSet("add", flag.ContinueOnError)
	description := flags.String("description", "", "entry description")
	amountText := flags.String("amount", "", "decimal amount; income positive, expense negative")
	dateText := flags.String("date", time.Now().Format(time.DateOnly), "entry date (YYYY-MM-DD)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	amount, err := money.Parse(*amountText)
	if err != nil {
		return err
	}
	date, err := time.Parse(time.DateOnly, *dateText)
	if err != nil {
		return fmt.Errorf("invalid date: use YYYY-MM-DD")
	}
	created, err := service.Create(context.Background(), bookkeeping.NewEntry{
		Description: *description,
		AmountCents: amount,
		OccurredAt:  date,
	})
	if err != nil {
		return err
	}
	fmt.Printf("added #%d %s %s\n", created.ID, money.Format(created.AmountCents), created.Description)
	return nil
}

func list(service *bookkeeping.Service, args []string) error {
	flags := flag.NewFlagSet("list", flag.ContinueOnError)
	if err := flags.Parse(args); err != nil {
		return err
	}
	entries, err := service.List(context.Background())
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDATE\tAMOUNT\tDESCRIPTION")
	for _, item := range entries {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", item.ID, item.OccurredAt.Format(time.DateOnly), money.Format(item.AmountCents), item.Description)
	}
	return w.Flush()
}

func usage() error {
	message := strings.TrimSpace(`market - small bookkeeping application

Usage:
  market migrate
  market serve [--addr :8080]
  market add --description TEXT --amount DECIMAL [--date YYYY-MM-DD]
  market list

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
