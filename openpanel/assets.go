// Package assets embeds static/ into the binary. A few files (custom.css, custom.js, robots.txt, security.txt) can be overridden by an on-disk copy at startup.
package assets

import "embed"

//go:embed all:static
var Static embed.FS
