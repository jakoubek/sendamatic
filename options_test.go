package sendamatic

import (
	"net/http"
	"testing"
	"time"
)

func TestWithBaseURL(t *testing.T) {
	customURL := "https://custom.api.url"
	client := NewClient("user", "pass", WithBaseURL(customURL))

	if client.baseURL != customURL {
		t.Errorf("baseURL = %q, want %q", client.baseURL, customURL)
	}
}

func TestWithTimeout(t *testing.T) {
	customTimeout := 60 * time.Second
	client := NewClient("user", "pass", WithTimeout(customTimeout))

	if client.httpClient.Timeout != customTimeout {
		t.Errorf("httpClient.Timeout = %v, want %v", client.httpClient.Timeout, customTimeout)
	}
}

func TestWithHTTPClient(t *testing.T) {
	customClient := &http.Client{
		Timeout: 90 * time.Second,
	}

	client := NewClient("user", "pass", WithHTTPClient(customClient))

	if client.httpClient != customClient {
		t.Error("httpClient not set to custom client")
	}

	if client.httpClient.Timeout != 90*time.Second {
		t.Errorf("httpClient.Timeout = %v, want 90s", client.httpClient.Timeout)
	}
}

func TestMultipleOptions(t *testing.T) {
	customURL := "https://test.api.url"
	customTimeout := 45 * time.Second

	client := NewClient("user", "pass",
		WithBaseURL(customURL),
		WithTimeout(customTimeout),
	)

	if client.baseURL != customURL {
		t.Errorf("baseURL = %q, want %q", client.baseURL, customURL)
	}

	if client.httpClient.Timeout != customTimeout {
		t.Errorf("httpClient.Timeout = %v, want %v", client.httpClient.Timeout, customTimeout)
	}
}

func TestDefaultValues(t *testing.T) {
	client := NewClient("user", "pass")

	if client.baseURL != defaultBaseURL {
		t.Errorf("baseURL = %q, want %q", client.baseURL, defaultBaseURL)
	}

	if client.httpClient.Timeout != defaultTimeout {
		t.Errorf("httpClient.Timeout = %v, want %v", client.httpClient.Timeout, defaultTimeout)
	}

	expectedAPIKey := "user-pass"
	if client.apiKey != expectedAPIKey {
		t.Errorf("apiKey = %q, want %q", client.apiKey, expectedAPIKey)
	}
}

func TestWithHTTPClient_PreservesCustomTransport(t *testing.T) {
	customTransport := &http.Transport{
		MaxIdleConns: 100,
	}

	customClient := &http.Client{
		Timeout:   45 * time.Second,
		Transport: customTransport,
	}

	client := NewClient("user", "pass", WithHTTPClient(customClient))

	if client.httpClient.Transport != customTransport {
		t.Error("Custom transport was not preserved")
	}
}

func TestOptionsOrder(t *testing.T) {
	// Test that options are applied in order
	// First set timeout to 30s, then provide a custom client with 60s
	customClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	client := NewClient("user", "pass",
		WithTimeout(30*time.Second),
		WithHTTPClient(customClient),
	)

	// The custom client should override the previous timeout setting
	if client.httpClient.Timeout != 60*time.Second {
		t.Errorf("httpClient.Timeout = %v, want 60s (custom client should override)", client.httpClient.Timeout)
	}
}

func TestWithTimeout_OverridesDefault(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{"short timeout", 5 * time.Second},
		{"long timeout", 120 * time.Second},
		{"zero timeout", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient("user", "pass", WithTimeout(tt.timeout))

			if client.httpClient.Timeout != tt.timeout {
				t.Errorf("httpClient.Timeout = %v, want %v", client.httpClient.Timeout, tt.timeout)
			}
		})
	}
}

func TestWithBaseURL_VariousURLs(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"production", "https://send.api.sendamatic.net"},
		{"staging", "https://staging.api.sendamatic.net"},
		{"local", "http://localhost:8080"},
		{"custom port", "https://api.example.com:8443"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient("user", "pass", WithBaseURL(tt.url))

			if client.baseURL != tt.url {
				t.Errorf("baseURL = %q, want %q", client.baseURL, tt.url)
			}
		})
	}
}
