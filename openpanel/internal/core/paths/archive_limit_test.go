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

func TestExtractStopsAtEntryLimit(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for i := 0; i < 10; i++ {
		_, _ = zw.Create("f" + string(rune('a'+i)))
	}
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

	ctx := WithSizeLimit(context.Background(), 1<<20)
	ctx.Value(sizeLimitKey{}).(*sizeBudget).entries.Store(5)
	if err := ExtractZip(ctx, root, ".", zr, 0); !errors.Is(err, ErrTooManyFiles) {
		t.Fatalf("err=%v, want ErrTooManyFiles", err)
	}
}

func TestExtractWithoutLimitGetsDefault(t *testing.T) {
	ctx := withDefaultLimit(context.Background())
	b, ok := ctx.Value(sizeLimitKey{}).(*sizeBudget)
	if !ok || b.left.Load() != DefaultExtractMaxBytes || b.entries.Load() != DefaultExtractMaxEntries {
		t.Fatal("an extraction without its own limit should get the default one")
	}
	if withDefaultLimit(WithSizeLimit(context.Background(), 5)).Value(sizeLimitKey{}).(*sizeBudget).left.Load() != 5 {
		t.Fatal("a caller's own limit must win")
	}
}
