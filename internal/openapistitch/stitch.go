package openapistitch

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/goccy/go-yaml"
)

type Config struct {
	SourceDir string
	Title     string
	Version   string
}

type pathOwner struct {
	route string
	file  string
}

var pathParameter = regexp.MustCompile(`\{[^}/]+\}`)

func Build(ctx context.Context, config Config) ([]byte, error) {
	if config.SourceDir == "" {
		return nil, fmt.Errorf("source directory is required")
	}
	if config.Title == "" {
		return nil, fmt.Errorf("API title is required")
	}
	if config.Version == "" {
		return nil, fmt.Errorf("API version is required")
	}

	files, err := filepath.Glob(filepath.Join(config.SourceDir, "*", "openapi.yaml"))
	if err != nil {
		return nil, fmt.Errorf("discover OpenAPI documents: %w", err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no OpenAPI documents found under %s", config.SourceDir)
	}

	components := newComponents()
	result := &openapi3.T{
		OpenAPI: "3.0.3",
		Info: &openapi3.Info{
			Title:   config.Title,
			Version: config.Version,
		},
		Paths:      openapi3.NewPaths(),
		Components: components,
	}
	componentOwners := make(map[string]string)
	pathOwners := make(map[string]pathOwner)
	operationOwners := make(map[string]string)

	for _, file := range files {
		loader := openapi3.NewLoader()
		loader.Context = ctx
		loader.IsExternalRefsAllowed = true
		doc, err := loader.LoadFromFile(file)
		if err != nil {
			return nil, fmt.Errorf("load %s: %w", file, err)
		}

		if doc.OpenAPI != "" {
			if err := doc.Validate(ctx); err != nil {
				return nil, fmt.Errorf("validate %s: %w", file, err)
			}
			if doc.OpenAPI != result.OpenAPI {
				return nil, fmt.Errorf("OpenAPI version mismatch in %s: got %s, want %s", file, doc.OpenAPI, result.OpenAPI)
			}
			for _, route := range doc.Paths.InMatchingOrder() {
				pathKey := pathParameter.ReplaceAllString(route, "{}")
				if previous, exists := pathOwners[pathKey]; exists {
					return nil, fmt.Errorf("path collision between %q in %s and %q in %s", previous.route, previous.file, route, file)
				}
				pathItem := doc.Paths.Find(route)
				for method, operation := range pathItem.Operations() {
					if operation.OperationID == "" {
						continue
					}
					location := method + " " + route + " in " + file
					if previous, exists := operationOwners[operation.OperationID]; exists {
						return nil, fmt.Errorf("operationId collision %q between %s and %s", operation.OperationID, previous, location)
					}
					operationOwners[operation.OperationID] = location
				}
				result.Paths.Set(route, pathItem)
				pathOwners[pathKey] = pathOwner{route: route, file: file}
			}
		}

		if doc.Components != nil {
			if err := mergeComponents(result.Components, doc.Components, file, componentOwners); err != nil {
				return nil, err
			}
		}
	}

	result.InternalizeRefs(ctx, componentName)
	if err := result.Validate(ctx); err != nil {
		return nil, fmt.Errorf("validate stitched OpenAPI document: %w", err)
	}

	output, err := yaml.MarshalWithOptions(result, yaml.Indent(2))
	if err != nil {
		return nil, fmt.Errorf("encode stitched OpenAPI document: %w", err)
	}
	return output, nil
}

func newComponents() *openapi3.Components {
	return &openapi3.Components{
		Schemas:         make(openapi3.Schemas),
		Parameters:      make(openapi3.ParametersMap),
		Headers:         make(openapi3.Headers),
		RequestBodies:   make(openapi3.RequestBodies),
		Responses:       make(openapi3.ResponseBodies),
		SecuritySchemes: make(openapi3.SecuritySchemes),
		Examples:        make(openapi3.Examples),
		Links:           make(openapi3.Links),
		Callbacks:       make(openapi3.Callbacks),
	}
}

func mergeComponents(dst, src *openapi3.Components, owner string, owners map[string]string) error {
	for _, merge := range []func() error{
		func() error { return mergeComponentMap(dst.Schemas, src.Schemas, "schemas", owner, owners) },
		func() error { return mergeComponentMap(dst.Parameters, src.Parameters, "parameters", owner, owners) },
		func() error { return mergeComponentMap(dst.Headers, src.Headers, "headers", owner, owners) },
		func() error {
			return mergeComponentMap(dst.RequestBodies, src.RequestBodies, "requestBodies", owner, owners)
		},
		func() error { return mergeComponentMap(dst.Responses, src.Responses, "responses", owner, owners) },
		func() error {
			return mergeComponentMap(dst.SecuritySchemes, src.SecuritySchemes, "securitySchemes", owner, owners)
		},
		func() error { return mergeComponentMap(dst.Examples, src.Examples, "examples", owner, owners) },
		func() error { return mergeComponentMap(dst.Links, src.Links, "links", owner, owners) },
		func() error { return mergeComponentMap(dst.Callbacks, src.Callbacks, "callbacks", owner, owners) },
	} {
		if err := merge(); err != nil {
			return err
		}
	}
	return nil
}

func mergeComponentMap[M ~map[string]V, V any](dst, src M, group, owner string, owners map[string]string) error {
	for name, value := range src {
		key := group + "/" + name
		if previous, exists := owners[key]; exists {
			return fmt.Errorf("component collision %q between %s and %s", key, previous, owner)
		}
		dst[name] = value
		owners[key] = owner
	}
	return nil
}

func componentName(doc *openapi3.T, ref openapi3.ComponentRef) string {
	if refPath := ref.RefPath(); refPath != nil {
		name := path.Base(strings.TrimPrefix(refPath.Fragment, "/"))
		if name != "." && name != "/" && name != "" {
			return name
		}
	}
	return openapi3.DefaultRefNameResolver(doc, ref)
}
