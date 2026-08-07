// Package assets embeds static/ (CSS/JS/vendor/images/flags) into the Go
// binary so deployment is a single file. A handful of paths under static/
// are meant to be user-editable
// after install (custom.css, custom.js, robots.txt, security.txt) - those
// are checked for a newer/overriding copy on disk at startup by
// cmd/openpanel, which prefers the disk version when present and falls
// back to the embedded default otherwise.
package assets

import "embed"

//go:embed all:static
var Static embed.FS
