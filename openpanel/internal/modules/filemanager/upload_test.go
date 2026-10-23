package filemanager

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
)

func TestParseUploadFormSpillsBigFilesToDisk(t *testing.T) {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile("files", "big.bin")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(make([]byte, uploadMemory+1<<20)); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodPost, "/x", &body)
	r.Header.Set("Content-Type", mw.FormDataContentType())

	if err := parseUploadForm(r); err != nil {
		t.Fatal(err)
	}
	defer r.MultipartForm.RemoveAll()
	f, err := r.MultipartForm.File["files"][0].Open()
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	// with the upload size limit as the budget a whole 2000 MB upload could sit in RAM
	if _, onDisk := f.(*os.File); !onDisk {
		t.Fatalf("a %d MB file part was kept in memory, want it in a temp file", (uploadMemory+1<<20)>>20)
	}
}

func TestUploadSucceeded(t *testing.T) {
	cases := []struct {
		name       string
		files      int
		allSuccess bool
		errs       []string
		want       bool
	}{
		{"all saved", 2, true, nil, true},
		// the form page said "uploaded successfully" next to "No files were uploaded"
		{"no files", 0, true, nil, false},
		// an invalid name adds an error without clearing allSuccess, the form page still said success
		{"rejected name", 1, true, []string{"Invalid filename: ..."}, false},
		{"failed save", 1, false, []string{"Failed to save a.txt"}, false},
	}
	for _, c := range cases {
		if got := uploadSucceeded(c.files, c.allSuccess, c.errs); got != c.want {
			t.Errorf("%s: uploadSucceeded = %v, want %v", c.name, got, c.want)
		}
	}
}

// the handlers defaulted to 5 MB while the pages and wget said 2000
func TestUploadSizeMB(t *testing.T) {
	cases := map[string]int{"": 2000, "300": 300, "junk": 2000}
	for raw, want := range cases {
		cfg := config.Config{}
		if raw != "" {
			cfg["filemanager_upload_size"] = raw
		}
		if got := uploadSizeMB(&appctx.App{Config: cfg}); got != want {
			t.Errorf("filemanager_upload_size=%q: got %d, want %d", raw, got, want)
		}
	}
}
