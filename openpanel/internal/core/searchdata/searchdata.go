// Package searchdata embeds the curated feature list used by the admin
// search feature. A compiled Go binary has no checked-out repo directory to
// read this file from at runtime, so the JSON is embedded directly rather
// than loaded from a path on disk, guaranteeing the search category is
// always populated.
package searchdata

import _ "embed"

//go:embed filter.json
var FeaturesJSON []byte
