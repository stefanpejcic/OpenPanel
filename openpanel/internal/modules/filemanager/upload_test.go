package filemanager

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

type readCounter struct {
	r io.Reader
	n int
}

func (c *readCounter) Read(p []byte) (int, error) {
	k, err := c.r.Read(p)
	c.n += k
	return k, err
}

// an upload over the limit used to be read in full and only then refused file by file
func TestReadUploadForm(t *testing.T) {
	form := func(sizes ...int) ([]byte, string) {
		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		for i, size := range sizes {
			fw, _ := mw.CreateFormFile("files", fmt.Sprintf("f%d.bin", i))
			_, _ = fw.Write(bytes.Repeat([]byte("x"), size))
		}
		_ = mw.Close()
		return buf.Bytes(), mw.FormDataContentType()
	}
	request := func(body []byte, ct string, length int64) (*http.Request, *readCounter) {
		rc := &readCounter{r: bytes.NewReader(body)}
		r := httptest.NewRequest(http.MethodPost, "/x", rc)
		r.ContentLength = length
		r.Header.Set("Content-Type", ct)
		return r, rc
	}
	// 1 MB per file makes the cap 2 MB plus the slack
	capBytes := 2<<20 + uploadSlack

	body, ct := form(8 << 20)
	r, rc := request(body, ct, int64(len(body)))
	if err := readUploadForm(httptest.NewRecorder(), r, 1); !errors.Is(err, errUploadTooLarge) || rc.n != 0 {
		t.Errorf("Content-Length over the cap: err %v after reading %d bytes, want errUploadTooLarge before reading", err, rc.n)
	}

	r, rc = request(body, ct, -1)
	if err := readUploadForm(httptest.NewRecorder(), r, 1); !errors.Is(err, errUploadTooLarge) || rc.n > capBytes+64<<10 {
		t.Errorf("no Content-Length: err %v after reading %d bytes, want errUploadTooLarge near %d", err, rc.n, capBytes)
	}

	// several files under the cap is the normal batch from the pages
	body, ct = form(900<<10, 900<<10)
	r, _ = request(body, ct, int64(len(body)))
	if err := readUploadForm(httptest.NewRecorder(), r, 1); err != nil || len(r.MultipartForm.File["files"]) != 2 {
		t.Fatalf("two 900 KB files: err %v", err)
	}
	if !strings.Contains(uploadTooLargeMessage(1), "2 MB in total and 1 MB per file") {
		t.Errorf("message: %s", uploadTooLargeMessage(1))
	}
}
