package docker

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
)

// ImageSummary is the containers page's Images tile
type ImageSummary struct {
	Count  int    `json:"count"`
	Unused int    `json:"unused"`
	Bytes  int64  `json:"bytes"`
	Size   string `json:"size"`
}

// parseImageList totals `podman image ls --format json`, an image no container uses counts as unused
func parseImageList(out []byte) (ImageSummary, error) {
	var images []struct {
		Size       int64 `json:"Size"`
		Containers int   `json:"Containers"`
	}
	if err := json.Unmarshal(out, &images); err != nil {
		return ImageSummary{}, err
	}
	var s ImageSummary
	for _, img := range images {
		s.Count++
		s.Bytes += img.Size
		if img.Containers == 0 {
			s.Unused++
		}
	}
	s.Size = decimalSize(s.Bytes)
	return s, nil
}

// parseSystemDF reads the Images row of `podman system df --format json`, its size counts layers shared between images once; ok is false when the row or RawSize is missing (older podman)
func parseSystemDF(out []byte) (s ImageSummary, ok bool) {
	var rows []struct {
		Type    string `json:"Type"`
		Total   int    `json:"Total"`
		Active  int    `json:"Active"`
		RawSize *int64 `json:"RawSize"`
	}
	if json.Unmarshal(out, &rows) != nil {
		return ImageSummary{}, false
	}
	for _, row := range rows {
		if row.Type != "Images" || row.RawSize == nil {
			continue
		}
		s = ImageSummary{Count: row.Total, Unused: max(row.Total-row.Active, 0), Bytes: *row.RawSize}
		s.Size = decimalSize(s.Bytes)
		return s, true
	}
	return ImageSummary{}, false
}

// decimalSize uses 1000-based units like podman's own SIZE column
func decimalSize(b int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	f := float64(b)
	i := 0
	for f >= 1000 && i < len(units)-1 {
		f /= 1000
		i++
	}
	if i == 0 {
		return strconv.FormatInt(b, 10) + " B"
	}
	return strconv.FormatFloat(f, 'f', 1, 64) + " " + units[i]
}

// listImages prefers `podman system df`, summing `podman image ls` would count shared layers once per image
func listImages(ctx context.Context, userContext string) (ImageSummary, error) {
	cctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if out, err := podmanmanager.Command(cctx, userContext, podmanmanager.PodmanArgv(userContext, "system", "df", "--format", "json")).Output(); err == nil {
		if s, ok := parseSystemDF(out); ok {
			return s, nil
		}
	}
	out, err := podmanmanager.Command(cctx, userContext, podmanmanager.PodmanArgv(userContext, "image", "ls", "--format", "json")).Output()
	if err != nil {
		return ImageSummary{}, err
	}
	return parseImageList(out)
}

// handleContainerImages serves GET /json/containers/images for the Images tile
func handleContainerImages(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, _, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	summary, err := listImages(r.Context(), userContext)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Could not list images")
		return
	}
	writeJSON(w, summary)
}
