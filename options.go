package sendamatic

import (
	"net/http"
	"time"
)

// Option is a function type that modifies a Client during initialization.
// Options follow the functional options pattern for configuring client behavior.
type Option func(*Client)

// WithBaseURL returns an Option that sets a custom API base URL for the client.
// Use this to point to a different Sendamatic API endpoint or a testing environment.
//
// Example:
//
//	client := sendamatic.NewClient("user", "pass",
//		sendamatic.WithBaseURL("https://custom.api.url"))
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithHTTPClient returns an Option that replaces the default HTTP client with a custom one.
// This allows full control over HTTP behavior such as transport settings, connection pooling,
// and custom middleware.
//
// Example:
//
//	customClient := &http.Client{
//		Timeout: 60 * time.Second,
//		Transport: customTransport,
//	}
//	client := sendamatic.NewClient("user", "pass",
//		sendamatic.WithHTTPClient(customClient))
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithTimeout returns an Option that sets the HTTP client timeout duration.
// This determines how long the client will wait for a response before timing out.
// The default timeout is 30 seconds.
//
// Example:
//
//	client := sendamatic.NewClient("user", "pass",
//		sendamatic.WithTimeout(60*time.Second))
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}
