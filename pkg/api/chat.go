package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

const defaultCompletionsURI = "https://api.githubcopilot.com/chat/completions"

// ChatResponseFunc is a function that is called for each chat response.
type ChatResponseFunc func(ChatResponse) error

// ChatCompletions sends a chat request to the GitHub Copilot API.
func (c *Client) ChatCompletions(ctx context.Context, token string, chatRequest *ChatRequest, fn ChatResponseFunc) error {
	f := func(bts []byte) error {
		var chatResponse ChatResponse
		if err := json.Unmarshal(bts, &chatResponse); err != nil {
			return err
		}
		return fn(chatResponse)
	}
	if chatRequest.N == 0 {
		chatRequest.N = 1
	}

	if chatRequest.Stream {
		return c.stream(ctx, token, defaultCompletionsURI, chatRequest, f)
	}
	return c.request(ctx, token, defaultCompletionsURI, chatRequest, f)
}

func (c *Client) doGet(ctx context.Context, token string, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	for key, value := range chatHeaders {
		req.Header.Set(key, value)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	return c.client.Do(req)
}

func (c *Client) doPost(ctx context.Context, token string, path string, data any) (*http.Response, error) {
	var buf *bytes.Buffer
	if data != nil {
		bts, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}

		buf = bytes.NewBuffer(bts)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path, buf)
	if err != nil {
		return nil, err
	}
	for key, value := range chatHeaders {
		req.Header.Set(key, value)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	return c.client.Do(req)
}

func (c *Client) stream(ctx context.Context, token string, path string, data any, fn func([]byte) error) error {
	response, err := c.doPost(ctx, token, path, data)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return statusError{
			StatusCode:   response.StatusCode,
			Status:       response.Status,
			ErrorMessage: string(body),
		}
	}

	scanner := bufio.NewScanner(response.Body)

	scanBuf := make([]byte, 0, 4096)
	scanner.Buffer(scanBuf, cap(scanBuf))
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return err
		}

		bts := scanner.Bytes()
		if !bytes.HasPrefix(bts, []byte("data:")) {
			continue
		}
		bts = bytes.TrimSpace(bts[5:])

		if bytes.Equal(bts, []byte("[DONE]")) {
			return nil
		}

		if err := fn(bts); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) request(ctx context.Context, token string, path string, data any, fn func([]byte) error) error {
	response, err := c.doPost(ctx, token, path, data)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return statusError{
			StatusCode:   response.StatusCode,
			Status:       response.Status,
			ErrorMessage: string(body),
		}
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if err := fn(body); err != nil {
		return err
	}
	return nil
}

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream,omitempty"`
	Temperature float64   `json:"temperature"`
	TopP        float64   `json:"top_p"`
	N           int64     `json:"n"`
}

type Message struct {
	Role    string `json:"role"` // one of ["system", "user", "assistant"]
	Content string `json:"content"`
}

type ChatResponse struct {
	Choices []Choice `json:"choices"`
	Created int64    `json:"created,omitempty"`
	ID      string   `json:"id"`
	Usage   Usage    `json:"usage,omitempty"`
}

type Usage struct {
	CompletionTokens int64 `json:"completion_tokens"`
	PromptTokens     int64 `json:"prompt_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
}

type Choice struct {
	FinishReason string  `json:"finish_reason,omitempty"`
	Index        int64   `json:"index,omitempty"`
	Message      Message `json:"message,omitempty"`
	Delta        Message `json:"delta,omitempty"`
}

type FilterResult struct {
	Filtered bool   `json:"filtered"`
	Severity string `json:"severity"`
}

var chatHeaders = map[string]string{
	"Editor-Version":        editorVersion,
	"Editor-Plugin-Version": editorPluginVersion,
	"User-Agent":            userAgent,
	"Openai-Organization":   openaiOrg,
	"Openai-Intent":         openaiIntent,
	"Content-Type":          "text/event-stream; charset=utf-8",
}
