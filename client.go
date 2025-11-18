package sendamatic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://send.api.sendamatic.net"
	defaultTimeout = 30 * time.Second
)

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewClient(userID, password string, opts ...Option) *Client {
	c := &Client{
		apiKey:  fmt.Sprintf("%s-%s", userID, password),
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}

	// Optionen anwenden
	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Send versendet eine E-Mail über die Sendamatic API
func (c *Client) Send(ctx context.Context, msg *Message) (*SendResponse, error) {
	if err := msg.Validate(); err != nil {
		return nil, fmt.Errorf("message validation failed: %w", err)
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/send", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Fehlerbehandlung für 4xx und 5xx
	if resp.StatusCode >= 400 {
		return nil, parseErrorResponse(resp.StatusCode, body)
	}

	var sendResp SendResponse
	if err := json.Unmarshal(body, &sendResp.Recipients); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	sendResp.StatusCode = resp.StatusCode
	return &sendResp, nil
}
