//go:build !gallery

package web

// normal builds never capture, so Render's gallery branch compiles away
const galleryEnabled = false

func captureGallery(paths []string, data any, body string) {}
