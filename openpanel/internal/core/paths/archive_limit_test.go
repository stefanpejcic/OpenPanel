package paths

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"os"
	"testing"
)

func TestExtractZipStopsAtSizeLimit(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create("zeros.bin")
	_, _ = w.Write(make([]byte, 2<<20))
	_ = zw.Close()
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	// a few KB of zip used to expand to whatever it held
	ctx := WithSizeLimit(context.Background(), 64<<10)
	if err := ExtractZip(ctx, root, ".", zr, 0); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("err=%v, want ErrTooLarge", err)
	}
	if err := ExtractZip(context.Background(), root, "ok", zr, 0); err != nil {
		t.Fatalf("no limit: %v", err)
	}
}
