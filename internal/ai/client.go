package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	base string
	http *http.Client
}

func NewClient(base string) *Client {
	return &Client{base: base, http: &http.Client{}}
}

type ChatMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatReq struct {
	Model       string    `json:"model"`
	Messages    []ChatMsg `json:"messages"`
	Stream      bool      `json:"stream"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

type chatStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// StreamChat sends messages and invokes onToken for each streamed token chunk.
// Returns the full accumulated assistant content.
func (c *Client) StreamChat(ctx context.Context, messages []ChatMsg, onToken func(string)) (string, error) {
	body, _ := json.Marshal(chatReq{
		Model:       "bit",
		Messages:    messages,
		Stream:      true,
		Temperature: 0.7,
		MaxTokens:   512,
	})
	req, err := http.NewRequestWithContext(ctx, "POST", c.base+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("chat %d: %s", resp.StatusCode, string(b))
	}

	var full strings.Builder
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			line = bytes.TrimSpace(line)
			if bytes.HasPrefix(line, []byte("data: ")) {
				payload := bytes.TrimPrefix(line, []byte("data: "))
				if bytes.Equal(payload, []byte("[DONE]")) {
					break
				}
				var chunk chatStreamChunk
				if jerr := json.Unmarshal(payload, &chunk); jerr == nil && len(chunk.Choices) > 0 {
					tok := chunk.Choices[0].Delta.Content
					if tok != "" {
						full.WriteString(tok)
						onToken(tok)
					}
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return full.String(), err
		}
		if ctx.Err() != nil {
			return full.String(), ctx.Err()
		}
	}
	return full.String(), nil
}
