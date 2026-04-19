package ai

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type ModelSpec struct {
	Name    string
	File    string
	URL     string
	Size    string
	SHA256  string // optional; empty skips verification
	RAMTier string
}

var Models = []ModelSpec{
	{
		Name:    "SmolLM2 360M Instruct (Q4_K_M)",
		File:    "smollm2-360m-instruct-q4_k_m.gguf",
		URL:     "https://huggingface.co/HuggingFaceTB/SmolLM2-360M-Instruct-GGUF/resolve/main/smollm2-360m-instruct-q4_k_m.gguf",
		Size:    "~230MB",
		RAMTier: "low",
	},
	{
		Name:    "Qwen2.5 0.5B Instruct (Q4_K_M)",
		File:    "qwen2.5-0.5b-instruct-q4_k_m.gguf",
		URL:     "https://huggingface.co/Qwen/Qwen2.5-0.5B-Instruct-GGUF/resolve/main/qwen2.5-0.5b-instruct-q4_k_m.gguf",
		Size:    "~390MB",
		RAMTier: "mid",
	},
	{
		Name:    "Phi-3.5 Mini Instruct (Q4_K_M)",
		File:    "Phi-3.5-mini-instruct-Q4_K_M.gguf",
		URL:     "https://huggingface.co/bartowski/Phi-3.5-mini-instruct-GGUF/resolve/main/Phi-3.5-mini-instruct-Q4_K_M.gguf",
		Size:    "~2.2GB",
		RAMTier: "high",
	},
}

// RecommendTier picks a model based on total RAM in GB.
func RecommendTier(ramGB int) int {
	switch {
	case ramGB < 4:
		return 0
	case ramGB < 8:
		return 1
	default:
		return 2
	}
}

// DownloadProgress is emitted during chunked downloads.
type DownloadProgress struct {
	Label    string
	Received int64
	Total    int64
	Done     bool
	Err      error
}

// DownloadFile downloads url to dst and streams progress on progress channel.
// Caller owns the channel (closed on completion).
func DownloadFile(url, dst, label string, progress chan<- DownloadProgress) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	tmp := dst + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer f.Close()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "bit-tracker/1.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	total := resp.ContentLength
	h := sha256.New()
	var received int64
	buf := make([]byte, 256*1024)
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				return werr
			}
			h.Write(buf[:n])
			received += int64(n)
			select {
			case progress <- DownloadProgress{Label: label, Received: received, Total: total}:
			default:
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return rerr
		}
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, dst); err != nil {
		return err
	}
	// emit done
	_ = hex.EncodeToString(h.Sum(nil))
	return nil
}

// LlamaServerRelease resolves the asset URL for the current GOOS/GOARCH.
// Uses a known-good pinned release to avoid hitting GitHub API.
func LlamaServerAssetURL() (string, error) {
	// Using llama.cpp release b4404 (Dec 2024). Caller may override via config.
	const base = "https://github.com/ggerganov/llama.cpp/releases/download/b4404"
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case "linux/amd64":
		return base + "/llama-b4404-bin-ubuntu-x64.zip", nil
	case "darwin/arm64":
		return base + "/llama-b4404-bin-macos-arm64.zip", nil
	case "darwin/amd64":
		return base + "/llama-b4404-bin-macos-x64.zip", nil
	case "windows/amd64":
		return base + "/llama-b4404-bin-win-avx2-x64.zip", nil
	}
	return "", fmt.Errorf("unsupported platform %s/%s", runtime.GOOS, runtime.GOARCH)
}

// ExtractLlamaServer extracts the `llama-server` binary from a downloaded zip.
func ExtractLlamaServer(zipPath, destBin string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()
	target := "llama-server"
	if runtime.GOOS == "windows" {
		target = "llama-server.exe"
	}
	for _, f := range r.File {
		base := filepath.Base(f.Name)
		if base == target {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			out, err := os.Create(destBin)
			if err != nil {
				rc.Close()
				return err
			}
			if _, err := io.Copy(out, rc); err != nil {
				rc.Close()
				out.Close()
				return err
			}
			rc.Close()
			out.Close()
			if runtime.GOOS != "windows" {
				_ = os.Chmod(destBin, 0o755)
			}
			return nil
		}
	}
	return errors.New("llama-server binary not found in archive")
}

// CleanFilename collapses whitespace/path separators.
func CleanFilename(s string) string {
	s = strings.ReplaceAll(s, string(filepath.Separator), "_")
	return s
}
