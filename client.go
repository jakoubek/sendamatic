// Package sendamatic provides a Go client library for the Sendamatic email delivery API.
//
// The library offers a simple and idiomatic Go API with context support, a fluent message
// builder interface, and comprehensive error handling for sending transactional emails.
//
// Example usage:
//
//	client := sendamatic.NewClient("your-user-id", "your-password")
//	msg := sendamatic.NewMessage().
//		SetSender("sender@example.com").
//		AddTo("recipient@example.com").
//		SetSubject("Hello").
//		SetTextBody("Hello World")
//	resp, err := client.Send(context.Background(), msg)
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
	// defaultBaseURL is the default Sendamatic API endpoint.
	defaultBaseURL = "https://send.api.sendamatic.net"
	// defaultTimeout is the default HTTP client timeout for API requests.
	defaultTimeout = 30 * time.Second
)

// Client represents a Sendamatic API client that handles authentication and HTTP communication
// with the Sendamatic email delivery service.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates and returns a new Client configured with the provided Sendamatic credentials.
// The userID and password are combined to form the API key used for authentication.
// Optional configuration functions can be provided to customize the client behavior.
//
// Example:
//
//	client := sendamatic.NewClient("user-id", "password",
//		sendamatic.WithTimeout(60*time.Second))
func NewClient(userID, password string, opts ...Option) *Client {
	c := &Client{
		apiKey:  fmt.Sprintf("%s-%s", userID, password),
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}

	// Apply configuration options
	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Send sends an email message through the Sendamatic API using the provided context.
// The message is validated before sending. If validation fails or the API request fails,
// an error is returned. On success, a SendResponse containing per-recipient delivery
// information is returned.
//
// The context can be used to set deadlines, timeouts, or cancel the request.
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
