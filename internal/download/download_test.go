package download

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFetchWritesFile(t *testing.T) {
	body := make([]byte, 64*1024)
	for i := range body {
		body[i] = byte(i)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}))
	defer srv.Close()

	dir := t.TempDir()
	dest := filepath.Join(dir, "sub", "blob.bin")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch := Fetch(ctx, srv.URL, dest)
	var last Progress
	for p := range ch {
		last = p
	}
	if last.Err != nil {
		t.Fatalf("err: %v", last.Err)
	}
	if !last.Done {
		t.Fatalf("no done event")
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(body) {
		t.Fatalf("short write: %d vs %d", len(got), len(body))
	}
}

func TestFetchHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	dir := t.TempDir()
	dest := filepath.Join(dir, "x.bin")
	ch := Fetch(context.Background(), srv.URL, dest)
	var last Progress
	for p := range ch {
		last = p
	}
	if last.Err == nil {
		t.Fatalf("expected error for 404")
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatalf("partial file should not remain")
	}
}

func TestHumanBytes(t *testing.T) {
	cases := map[int64]string{
		10:           "10 B",
		2048:         "2 KB",
		3 * 1 << 20:  "3 MB",
		2 * 1 << 30:  "2.0 GB",
	}
	for in, want := range cases {
		if got := HumanBytes(in); got != want {
			t.Errorf("HumanBytes(%d) = %q, want %q", in, got, want)
		}
	}
}
