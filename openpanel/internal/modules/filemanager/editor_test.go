package filemanager

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
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

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("disk read failed") }

// a download cut off by the time limit or a read error used to end as a clean, complete looking response
func TestSendFile(t *testing.T) {
	data := strings.Repeat("x", 3*8192)
	size := int64(len(data))
	aborted := func(t *testing.T, fn func()) {
		t.Helper()
		defer func() {
			if r := recover(); r != http.ErrAbortHandler {
				t.Errorf("got panic %v, want http.ErrAbortHandler", r)
			}
		}()
		fn()
	}

	w := httptest.NewRecorder()
	sendFile(w, strings.NewReader(data), size, func() bool { return false })
	if w.Body.Len() != len(data) || w.Header().Get("Content-Length") != strconv.FormatInt(size, 10) {
		t.Fatalf("whole file: sent %d bytes, Content-Length %q", w.Body.Len(), w.Header().Get("Content-Length"))
	}

	calls := 0
	w = httptest.NewRecorder()
	aborted(t, func() { sendFile(w, strings.NewReader(data), size, func() bool { calls++; return calls > 1 }) })
	if w.Body.Len() >= len(data) {
		t.Errorf("time limit: the whole file went out")
	}

	w = httptest.NewRecorder()
	aborted(t, func() {
		sendFile(w, io.MultiReader(strings.NewReader(data[:8192]), failingReader{}), size, func() bool { return false })
	})

	// the file got shorter after its size was taken
	w = httptest.NewRecorder()
	aborted(t, func() { sendFile(w, strings.NewReader(data[:100]), size, func() bool { return false }) })
}
