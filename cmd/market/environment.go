package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func loadDevelopmentEnv() error {
	if envOr("ENVIRONMENT", "DEVELOPMENT") != "DEVELOPMENT" {
		return nil
	}
	// Load preserves existing process variables, including explicitly empty values.
	if err := godotenv.Load(".env"); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("load development .env: %w", err)
	}
	return nil
}
