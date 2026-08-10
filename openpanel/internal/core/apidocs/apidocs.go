// Package apidocs holds shared endpoint metadata that is the
// source-of-truth JSON for both the app's own MCP tool descriptions and
// website/scripts/generate_openapi.py's OpenAPI spec generation.
//
// The data is pure documentation - group/method/path/description/example
// body - kept here as pre-generated JSON rather than re-typed as Go struct
// literals, eliminating any risk of a transcription mismatch.
package apidocs

import _ "embed"

//go:embed api_endpoints.json
var EndpointsJSON []byte
