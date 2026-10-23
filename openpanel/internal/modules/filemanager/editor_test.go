package filemanager

import (
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
)

func TestEditSizeMBReadsTheConfiguredKey(t *testing.T) {
	// the editor used to read filemanager_editor_size, a key nothing ever sets
	a := &appctx.App{Config: config.Config{"filemanager_edit_size": "7", "filemanager_editor_size": "3"}}
	if got := editSizeMB(a); got != 7 {
		t.Fatalf("editSizeMB = %d, want 7 from filemanager_edit_size", got)
	}
	if got := editSizeMB(&appctx.App{Config: config.Config{}}); got != 5 {
		t.Fatalf("editSizeMB default = %d, want 5", got)
	}
}
