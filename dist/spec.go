// Package apidoc exposes the complete OpenAPI contract embedded in the binary.
package apidoc

import _ "embed"

// OpenAPI is the stitched OpenAPI document.
//
//go:embed openapi.yaml
var OpenAPI []byte
