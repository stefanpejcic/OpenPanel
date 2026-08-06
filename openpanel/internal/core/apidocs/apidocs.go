// Package apidocs holds shared endpoint metadata used by the /account/api
// docs page, and doubles as the source-of-truth JSON for the app's own MCP
// tool descriptions.
//
// The data is pure documentation - group/method/path/description/example
// body - consumed entirely client-side by api_docs.html's Alpine
// component, so it's kept here as pre-generated JSON rather than re-typed
// as Go struct literals, eliminating any risk of a transcription mismatch.
package apidocs

import _ "embed"

//go:embed api_endpoints.json
var EndpointsJSON []byte
