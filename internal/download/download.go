// Package download streams a file over HTTP with progress reporting.
//
// It writes to a `.part` file, then atomically renames on success.
// Progress is delivered via a channel of Progress events so the Bubble Tea
// layer can turn them into tea.Msgs at its own cadence.
package download

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Progress describes the current state of a download.
type Progress struct {
	URL        string
	Dest       string
	BytesDone  int64
	BytesTotal int64 // 0 if unknown
	Started    time.Time
	Done       bool
	Err        error
}

// Percent returns 0..100, or -1 when total is unknown.
func (p Progress) Percent() int {
	if p.BytesTotal <= 0 {
		return -1
	}
	return int(p.BytesDone * 100 / p.BytesTotal)
}

// Speed returns bytes/sec averaged from Started.
func (p Progress) Speed() float64 {
	dt := time.Since(p.Started).Seconds()
	if dt <= 0 {
		return 0
	}
	return float64(p.BytesDone) / dt
}

// ETA returns seconds remaining, or -1 when unknown.
func (p Progress) ETA() int {
	if p.BytesTotal <= 0 {
		return -1
	}
	sp := p.Speed()
	if sp <= 0 {
		return -1
	}
	remain := p.BytesTotal - p.BytesDone
	if remain < 0 {
		return 0
	}
	return int(float64(remain) / sp)
}

// Fetch downloads url to dest, reporting progress on the returned channel.
// dest's parent directory is created if missing.
// If dest already exists and size matches the Content-Length, the download
// is skipped and a single Done progress is emitted.
func Fetch(ctx context.Context, url, dest string) <-chan Progress {
	ch := make(chan Progress, 8)
	go func() {
		defer close(ch)
		started := time.Now()
		emit := func(p Progress) {
			p.URL = url
			p.Dest = dest
			p.Started = started
			select {
			case ch <- p:
			case <-ctx.Done():
			}
		}

		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			emit(Progress{Err: err, Done: true})
			return
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			emit(Progress{Err: err, Done: true})
			return
		}
		req.Header.Set("User-Agent", "bit-tracker/0.1 (+https://github.com/andre-cmd-rgb/bit-tracker)")

		client := &http.Client{Timeout: 0} // long downloads; ctx governs cancellation
		resp, err := client.Do(req)
		if err != nil {
			emit(Progress{Err: err, Done: true})
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			emit(Progress{Err: fmt.Errorf("http %s", resp.Status), Done: true})
			return
		}
		total := resp.ContentLength

		// Skip if target already exists with matching size.
		if info, err := os.Stat(dest); err == nil && total > 0 && info.Size() == total {
			emit(Progress{BytesDone: total, BytesTotal: total, Done: true})
			return
		}

		tmp := dest + ".part"
		out, err := os.Create(tmp)
		if err != nil {
			emit(Progress{Err: err, Done: true})
			return
		}

		buf := make([]byte, 256*1024)
		var done int64
		lastEmit := time.Now()
		for {
			if err := ctx.Err(); err != nil {
				out.Close()
				os.Remove(tmp)
				emit(Progress{BytesDone: done, BytesTotal: total, Err: err, Done: true})
				return
			}
			n, rerr := resp.Body.Read(buf)
			if n > 0 {
				if _, werr := out.Write(buf[:n]); werr != nil {
					out.Close()
					os.Remove(tmp)
					emit(Progress{BytesDone: done, BytesTotal: total, Err: werr, Done: true})
					return
				}
				done += int64(n)
				if time.Since(lastEmit) >= 200*time.Millisecond {
					emit(Progress{BytesDone: done, BytesTotal: total})
					lastEmit = time.Now()
				}
			}
			if rerr != nil {
				if errors.Is(rerr, io.EOF) {
					break
				}
				out.Close()
				os.Remove(tmp)
				emit(Progress{BytesDone: done, BytesTotal: total, Err: rerr, Done: true})
				return
			}
		}
		if err := out.Close(); err != nil {
			os.Remove(tmp)
			emit(Progress{BytesDone: done, BytesTotal: total, Err: err, Done: true})
			return
		}
		if err := os.Rename(tmp, dest); err != nil {
			emit(Progress{BytesDone: done, BytesTotal: total, Err: err, Done: true})
			return
		}
		emit(Progress{BytesDone: done, BytesTotal: total, Done: true})
	}()
	return ch
}

// HumanBytes formats bytes as a short human string ("123 MB", "1.4 GB").
func HumanBytes(n int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
	)
	switch {
	case n >= gb:
		return fmt.Sprintf("%.1f GB", float64(n)/float64(gb))
	case n >= mb:
		return fmt.Sprintf("%.0f MB", float64(n)/float64(mb))
	case n >= kb:
		return fmt.Sprintf("%.0f KB", float64(n)/float64(kb))
	default:
		return fmt.Sprintf("%d B", n)
	}
}
