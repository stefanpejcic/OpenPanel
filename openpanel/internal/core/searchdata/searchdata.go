// Package searchdata embeds the curated feature list for the admin search feature, since a compiled binary has no repo directory to read it from on disk.
package searchdata

import _ "embed"

//go:embed filter.json
var FeaturesJSON []byte
