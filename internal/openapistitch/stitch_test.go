package openapistitch_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dvtkrlbs/market/internal/openapistitch"
	"github.com/getkin/kin-openapi/openapi3"
)

func TestBuildCombinesModulesAndInternalizesSharedComponents(t *testing.T) {
	sourceDir := t.TempDir()
	writeSpec(t, sourceDir, "common", `
components:
  schemas:
    Error:
      type: object
      required: [message]
      properties:
        message: {type: string}
`)
	writeSpec(t, sourceDir, "health", `
openapi: 3.0.3
info: {title: Health, version: 0.1.0}
paths:
  /health:
    get:
      operationId: getHealth
      responses:
        "200": {description: Healthy}
`)
	writeSpec(t, sourceDir, "entries", `
openapi: 3.0.3
info: {title: Entries, version: 0.1.0}
paths:
  /entries:
    post:
      operationId: createEntry
      responses:
        "400":
          description: Invalid request
          content:
            application/json:
              schema:
                $ref: "../common/openapi.yaml#/components/schemas/Error"
`)

	result, err := openapistitch.Build(context.Background(), openapistitch.Config{
		SourceDir: sourceDir,
		Title:     "Market API",
		Version:   "0.1.0",
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if strings.Contains(string(result), "../common/openapi.yaml") {
		t.Fatalf("Build() left an external reference:\n%s", result)
	}
	if !strings.Contains(string(result), "paths:\n  /entries:") {
		t.Fatalf("Build() did not use two-space YAML indentation:\n%s", result)
	}

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(result)
	if err != nil {
		t.Fatalf("load result: %v", err)
	}
	if err := doc.Validate(context.Background()); err != nil {
		t.Fatalf("validate result: %v", err)
	}
	if doc.Info.Title != "Market API" || doc.Info.Version != "0.1.0" {
		t.Fatalf("info = %q %q, want Market API 0.1.0", doc.Info.Title, doc.Info.Version)
	}
	if doc.Paths.Find("/health") == nil || doc.Paths.Find("/entries") == nil {
		t.Fatalf("paths do not contain both module routes: %v", doc.Paths.InMatchingOrder())
	}
	if doc.Components == nil || doc.Components.Schemas["Error"] == nil {
		t.Fatalf("components do not contain shared Error schema")
	}
}

func TestBuildRejectsPathCollisions(t *testing.T) {
	sourceDir := t.TempDir()
	writeSpec(t, sourceDir, "entries", `
openapi: 3.0.3
info: {title: Module, version: 0.1.0}
paths:
  /users/{userId}/items:
    get:
      operationId: getUserItems
      parameters:
        - name: userId
          in: path
          required: true
          schema: {type: string}
      responses:
        "200": {description: Items}
`)
	writeSpec(t, sourceDir, "orders", `
openapi: 3.0.3
info: {title: Module, version: 0.1.0}
paths:
  /users/{id}/items:
    get:
      operationId: getItemsByID
      parameters:
        - name: id
          in: path
          required: true
          schema: {type: string}
      responses:
        "200": {description: Items}
`)

	_, err := openapistitch.Build(context.Background(), openapistitch.Config{
		SourceDir: sourceDir,
		Title:     "Market API",
		Version:   "0.1.0",
	})
	if err == nil || !strings.Contains(err.Error(), `path collision`) {
		t.Fatalf("Build() error = %v, want path collision", err)
	}
	if !strings.Contains(err.Error(), "entries") || !strings.Contains(err.Error(), "orders") {
		t.Fatalf("Build() error = %v, want both module names", err)
	}
}

func TestBuildRejectsOperationIDCollisions(t *testing.T) {
	sourceDir := t.TempDir()
	writeSpec(t, sourceDir, "entries", `
openapi: 3.0.3
info: {title: Entries, version: 0.1.0}
paths:
  /entries:
    get:
      operationId: listItems
      responses:
        "200": {description: Entries}
`)
	writeSpec(t, sourceDir, "orders", `
openapi: 3.0.3
info: {title: Orders, version: 0.1.0}
paths:
  /orders:
    get:
      operationId: listItems
      responses:
        "200": {description: Orders}
`)

	_, err := openapistitch.Build(context.Background(), openapistitch.Config{
		SourceDir: sourceDir,
		Title:     "Market API",
		Version:   "0.1.0",
	})
	if err == nil || !strings.Contains(err.Error(), `operationId collision "listItems"`) {
		t.Fatalf("Build() error = %v, want operationId collision", err)
	}
	if !strings.Contains(err.Error(), "GET /entries") || !strings.Contains(err.Error(), "GET /orders") {
		t.Fatalf("Build() error = %v, want both operations", err)
	}
}

func TestBuildRejectsComponentCollisions(t *testing.T) {
	sourceDir := t.TempDir()
	for _, module := range []string{"entries", "orders"} {
		writeSpec(t, sourceDir, module, `
openapi: 3.0.3
info: {title: Module, version: 0.1.0}
paths: {}
components:
  schemas:
    Item:
      type: object
`)
	}

	_, err := openapistitch.Build(context.Background(), openapistitch.Config{
		SourceDir: sourceDir,
		Title:     "Market API",
		Version:   "0.1.0",
	})
	if err == nil || !strings.Contains(err.Error(), `component collision "schemas/Item"`) {
		t.Fatalf("Build() error = %v, want component collision", err)
	}
	if !strings.Contains(err.Error(), "entries") || !strings.Contains(err.Error(), "orders") {
		t.Fatalf("Build() error = %v, want both module names", err)
	}
}

func TestBuildDeduplicatesSharedSecuritySchemes(t *testing.T) {
	sourceDir := t.TempDir()
	writeSpec(t, sourceDir, "common", `
components:
  securitySchemes:
    BearerAuth: {type: http, scheme: bearer}
`)
	writeSpec(t, sourceDir, "auth", `
openapi: 3.0.3
info: {title: Auth, version: 0.1.0}
paths:
  /whoami:
    get:
      security: [{BearerAuth: []}]
      responses:
        "200": {description: Authenticated}
components:
  securitySchemes:
    BearerAuth:
      $ref: "../common/openapi.yaml#/components/securitySchemes/BearerAuth"
`)

	result, err := openapistitch.Build(context.Background(), openapistitch.Config{
		SourceDir: sourceDir,
		Title:     "Market API",
		Version:   "0.1.0",
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(result)
	if err != nil {
		t.Fatalf("load result: %v", err)
	}
	if got := doc.Components.SecuritySchemes["BearerAuth"]; got == nil || got.Value == nil || got.Value.Scheme != "bearer" {
		t.Fatalf("BearerAuth = %#v, want bearer scheme", got)
	}
}

func writeSpec(t *testing.T, root, module, contents string) {
	t.Helper()
	dir := filepath.Join(root, module)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "openapi.yaml"), []byte(strings.TrimSpace(contents)+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}
