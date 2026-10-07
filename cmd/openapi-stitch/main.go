package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dvtkrlbs/market/internal/openapistitch"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("openapi-stitch", flag.ContinueOnError)
	source := flags.String("source", "api", "directory containing OpenAPI modules")
	output := flags.String("output", "dist/openapi.yaml", "bundled OpenAPI output path")
	title := flags.String("title", "Market API", "title for the complete API")
	version := flags.String("version", "0.1.0", "version for the complete API")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("positional arguments are not accepted")
	}

	contents, err := openapistitch.Build(ctx, openapistitch.Config{
		SourceDir: *source,
		Title:     *title,
		Version:   *version,
	})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(*output), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	if err := os.WriteFile(*output, contents, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", *output, err)
	}
	return nil
}
