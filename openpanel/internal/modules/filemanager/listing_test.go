package filemanager

import (
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
)

func TestFilesPerPageNeverZero(t *testing.T) {
	// 0 passes OpenAdmin's non-negative check and used to crash every listing with a division by zero
	for value, want := range map[string]int{"0": 500, "-5": 500, "": 500, "abc": 500, "50": 50, "1000": 1000} {
		a := &appctx.App{Config: config.Config{"filemanager_files_per_page": value}}
		if got := filesPerPage(a); got != want {
			t.Errorf("filesPerPage(%q) = %d, want %d", value, got, want)
		}
	}
}
