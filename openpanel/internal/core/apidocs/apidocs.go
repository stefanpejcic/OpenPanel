// Package apidocs holds the endpoint metadata JSON that's the source of truth for both the MCP tool descriptions and website/scripts/generate_openapi.py's OpenAPI spec.
package apidocs

import _ "embed"

//go:embed api_endpoints.json
var EndpointsJSON []byte
