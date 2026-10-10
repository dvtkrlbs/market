package main

import (
	"os"
	"path/filepath"
	"testing"
)

const testEnvVariable = "MARKET_DOTENV_TEST_VALUE"

func setupEnvironmentTest(t *testing.T, environment, contents string) {
	t.Helper()
	t.Setenv("ENVIRONMENT", environment)
	t.Setenv(testEnvVariable, "")
	if err := os.Unsetenv(testEnvVariable); err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	t.Chdir(directory)
	if err := os.WriteFile(filepath.Join(directory, ".env"), []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestLoadDevelopmentEnv(t *testing.T) {
	for _, environment := range []string{"", "DEVELOPMENT"} {
		t.Run(environment, func(t *testing.T) {
			setupEnvironmentTest(t, environment, testEnvVariable+"=from-file\n")
			if err := loadDevelopmentEnv(); err != nil {
				t.Fatal(err)
			}
			if got := os.Getenv(testEnvVariable); got != "from-file" {
				t.Fatalf("dotenv value = %q, want from-file", got)
			}
		})
	}
}

func TestLoadDevelopmentEnvPreservesProcessValues(t *testing.T) {
	for _, value := range []string{"from-process", ""} {
		t.Run(value, func(t *testing.T) {
			setupEnvironmentTest(t, "DEVELOPMENT", testEnvVariable+"=from-file\n")
			t.Setenv(testEnvVariable, value)
			if err := loadDevelopmentEnv(); err != nil {
				t.Fatal(err)
			}
			if got := os.Getenv(testEnvVariable); got != value {
				t.Fatalf("dotenv overwrote process value: got %q, want %q", got, value)
			}
		})
	}
}

func TestLoadDevelopmentEnvSkipsOtherEnvironments(t *testing.T) {
	for _, environment := range []string{"PRODUCTION", "STAGING"} {
		t.Run(environment, func(t *testing.T) {
			setupEnvironmentTest(t, environment, "ENVIRONMENT=DEVELOPMENT\n"+testEnvVariable+"=from-file\n")
			if err := loadDevelopmentEnv(); err != nil {
				t.Fatal(err)
			}
			if _, exists := os.LookupEnv(testEnvVariable); exists {
				t.Fatal("loaded .env outside development")
			}
		})
	}
}

func TestLoadDevelopmentEnvMissingFile(t *testing.T) {
	t.Setenv("ENVIRONMENT", "DEVELOPMENT")
	t.Chdir(t.TempDir())
	if err := loadDevelopmentEnv(); err != nil {
		t.Fatalf("missing .env error = %v, want nil", err)
	}
}

func TestLoadDevelopmentEnvMalformedFile(t *testing.T) {
	setupEnvironmentTest(t, "DEVELOPMENT", testEnvVariable+"=\"unterminated\n")
	if err := loadDevelopmentEnv(); err == nil {
		t.Fatal("expected malformed dotenv error")
	}
}
