package diskusage

import (
	"context"
	"errors"
	"net/http"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
)

const (
	largestFilesLimit   = 20
	largestFilesTimeout = 10 * time.Second
	largestFilesTTL     = time.Minute
)

// LargeFile is one row of the Largest files card
type LargeFile struct {
	Path            string `json:"path"`
	Bytes           int64  `json:"bytes"`
	Size            string `json:"size"`
	Pct             int    `json:"pct"`
	FileManagerHref string `json:"file_manager_href,omitempty"`
}

// largestFilesResult is what ?largest=1 returns, Partial is set when find hit the time limit
type largestFilesResult struct {
	Files   []LargeFile `json:"files"`
	Partial bool        `json:"partial"`
}

// humanBytes matches du -h's binary units, as the folder list shows them
func humanBytes(b int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	f := float64(b)
	i := 0
	for f >= 1024 && i < len(units)-1 {
		f /= 1024
		i++
	}
	if i == 0 {
		return strconv.FormatInt(b, 10) + " B"
	}
	if f >= 100 {
		return strconv.FormatFloat(f, 'f', 0, 64) + " " + units[i]
	}
	return strconv.FormatFloat(f, 'f', 1, 64) + " " + units[i]
}

// imageStoragePrune is the podman image store under the folder, those layers show on the Containers page instead
func imageStoragePrune(rel string) string {
	switch rel {
	case ".":
		return "./docker-data/overlay*"
	case "docker-data":
		return "./overlay*"
	}
	return ""
}

// parseLargestFiles turns `find -printf '%s\t%P\n'` lines into the top n files, biggest first
func parseLargestFiles(out, urlPath string, n int) []LargeFile {
	var files []LargeFile
	for _, line := range strings.Split(out, "\n") {
		sizeText, name, ok := strings.Cut(line, "\t")
		if !ok || name == "" {
			continue
		}
		size, err := strconv.ParseInt(sizeText, 10, 64)
		if err != nil {
			continue
		}
		files = append(files, LargeFile{Path: name, Bytes: size})
	}
	sort.SliceStable(files, func(i, j int) bool { return files[i].Bytes > files[j].Bytes })
	if len(files) > n {
		files = files[:n]
	}
	for i := range files {
		files[i].Size = humanBytes(files[i].Bytes)
		files[i].Pct = max(1, int(files[i].Bytes*100/max(1, files[0].Bytes)))
		dir := path.Dir(files[i].Path)
		if dir == "." {
			dir = ""
		} else {
			dir += "/"
		}
		if fm, ok := fileManagerBase(escapePath(urlPath + dir)); ok {
			files[i].FileManagerHref = strings.TrimSuffix(fm, "/")
			if files[i].FileManagerHref == "/files" {
				files[i].FileManagerHref = "/files/"
			}
		}
	}
	return files
}

// largestFiles runs a time-limited find for files over 1 MB under actualFolder, cached a minute
func largestFiles(ctx context.Context, a *appctx.App, userContext, actualFolder, urlPath string) largestFilesResult {
	res, _ := cache.Memoize(ctx, a.Cache, "disk_usage_largest:"+userContext+":"+actualFolder, largestFilesTTL, func() (largestFilesResult, error) {
		home := "/home/" + userContext
		rel, err := filepath.Rel(home, actualFolder)
		if err != nil {
			return largestFilesResult{}, nil //nolint:nilerr // unreachable after resolveUnderHome, show nothing
		}
		args := []string{".", "-xdev"}
		if p := imageStoragePrune(filepath.ToSlash(rel)); p != "" {
			args = append(args, "-path", p, "-prune", "-o")
		}
		args = append(args, "-type", "f", "-size", "+1M", "-printf", `%s\t%P\n`)

		cctx, cancel := context.WithTimeout(ctx, largestFilesTimeout)
		defer cancel()
		out, err := tenantfs.OutputIn(cctx, home, actualFolder, "find", args...)
		// permission errors still leave the rest of the output, a timeout leaves what was found so far
		partial := errors.Is(cctx.Err(), context.DeadlineExceeded)
		if err != nil && len(out) == 0 && !partial {
			return largestFilesResult{}, nil //nolint:nilerr // command failure -> empty card, not an error response
		}
		return largestFilesResult{Files: parseLargestFiles(string(out), urlPath, largestFilesLimit), Partial: partial}, nil
	})
	if res.Files == nil {
		res.Files = []LargeFile{}
	}
	return res
}

func handleLargestFiles(ctx context.Context, a *appctx.App, w http.ResponseWriter, userContext, actualFolder, urlPath string) {
	writeJSON(w, largestFiles(ctx, a, userContext, actualFolder, urlPath))
}
