package sendamatic

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-user", "test-pass")

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	expectedAPIKey := "test-user-test-pass"
	if client.apiKey != expectedAPIKey {
		t.Errorf("apiKey = %q, want %q", client.apiKey, expectedAPIKey)
	}

	if client.baseURL != defaultBaseURL {
		t.Errorf("baseURL = %q, want %q", client.baseURL, defaultBaseURL)
	}

	if client.httpClient == nil {
		t.Fatal("httpClient is nil")
	}

	if client.httpClient.Timeout != defaultTimeout {
		t.Errorf("httpClient.Timeout = %v, want %v", client.httpClient.Timeout, defaultTimeout)
	}
}

func TestClient_Send_Success(t *testing.T) {
	// Create a test server that returns a successful response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != http.MethodPost {
			t.Errorf("Method = %s, want POST", r.Method)
		}

		// Verify request path
		if r.URL.Path != "/send" {
			t.Errorf("Path = %s, want /send", r.URL.Path)
		}

		// Verify headers
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %s, want application/json", ct)
		}

		if apiKey := r.Header.Get("x-api-key"); apiKey != "user-pass" {
			t.Errorf("x-api-key = %s, want user-pass", apiKey)
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var msg Message
		if err := json.Unmarshal(body, &msg); err != nil {
			t.Errorf("Failed to unmarshal request body: %v", err)
		}

		// Send successful response
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		response := map[string][2]interface{}{
			"recipient@example.com": {float64(200), "msg-12345"},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("user", "pass", WithBaseURL(server.URL))

	msg := NewMessage().
		SetSender("sender@example.com").
		AddTo("recipient@example.com").
		SetSubject("Test").
		SetTextBody("Body")

	resp, err := client.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("Send() error = %v, want nil", err)
	}

	if !resp.IsSuccess() {
		t.Error("Expected successful response")
	}

	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}

	msgID, ok := resp.GetMessageID("recipient@example.com")
	if !ok {
		t.Error("Expected to find message ID")
	}
	if msgID != "msg-12345" {
		t.Errorf("MessageID = %q, want %q", msgID, "msg-12345")
	}
}

func TestClient_Send_ValidationError(t *testing.T) {
	client := NewClient("user", "pass")

	// Create an invalid message (no recipients)
	msg := NewMessage().
		SetSender("sender@example.com").
		SetSubject("Test").
		SetTextBody("Body")

	_, err := client.Send(context.Background(), msg)
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("Error message = %q, want to contain 'validation failed'", err.Error())
	}
}

func TestClient_Send_APIError(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		wantErrMessage string
	}{
		{
			name:           "400 bad request",
			statusCode:     400,
			responseBody:   `{"error": "Invalid request"}`,
			wantErrMessage: "sendamatic api error (status 400): Invalid request",
		},
		{
			name:           "401 unauthorized",
			statusCode:     401,
			responseBody:   `{"error": "Invalid API key"}`,
			wantErrMessage: "sendamatic api error (status 401): Invalid API key",
		},
		{
			name:           "422 validation error",
			statusCode:     422,
			responseBody:   `{"error": "Validation failed", "validation_errors": "sender is required", "json_path": "$.sender"}`,
			wantErrMessage: "sendamatic api error (status 422): sender is required (path: $.sender)",
		},
		{
			name:           "500 server error",
			statusCode:     500,
			responseBody:   `{"error": "Internal server error"}`,
			wantErrMessage: "sendamatic api error (status 500): Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewClient("user", "pass", WithBaseURL(server.URL))

			msg := NewMessage().
				SetSender("sender@example.com").
				AddTo("recipient@example.com").
				SetSubject("Test").
				SetTextBody("Body")

			_, err := client.Send(context.Background(), msg)
			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("Error type = %T, want *APIError", err)
			}

			if apiErr.StatusCode != tt.statusCode {
				t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.statusCode)
			}

			if err.Error() != tt.wantErrMessage {
				t.Errorf("Error message = %q, want %q", err.Error(), tt.wantErrMessage)
			}
		})
	}
}

func TestClient_Send_ContextTimeout(t *testing.T) {
	// Create a server that delays the response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"recipient@example.com": [200, "msg-12345"]}`))
	}))
	defer server.Close()

	client := NewClient("user", "pass", WithBaseURL(server.URL))

	msg := NewMessage().
		SetSender("sender@example.com").
		AddTo("recipient@example.com").
		SetSubject("Test").
		SetTextBody("Body")

	// Create a context that times out quickly
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.Send(ctx, msg)
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}

	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Errorf("Expected context deadline exceeded error, got: %v", err)
	}
}

func TestClient_Send_ContextCancellation(t *testing.T) {
	// Create a server that delays the response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"recipient@example.com": [200, "msg-12345"]}`))
	}))
	defer server.Close()

	client := NewClient("user", "pass", WithBaseURL(server.URL))

	msg := NewMessage().
		SetSender("sender@example.com").
		AddTo("recipient@example.com").
		SetSubject("Test").
		SetTextBody("Body")

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the context immediately
	cancel()

	_, err := client.Send(ctx, msg)
	if err == nil {
		t.Fatal("Expected cancellation error, got nil")
	}

	if !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("Expected context canceled error, got: %v", err)
	}
}

func TestClient_Send_MultipleRecipients(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		response := map[string][2]interface{}{
			"recipient1@example.com": {float64(200), "msg-11111"},
			"recipient2@example.com": {float64(200), "msg-22222"},
			"recipient3@example.com": {float64(550), "msg-33333"}, // Failed delivery
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("user", "pass", WithBaseURL(server.URL))

	msg := NewMessage().
		SetSender("sender@example.com").
		AddTo("recipient1@example.com").
		AddTo("recipient2@example.com").
		AddTo("recipient3@example.com").
		SetSubject("Test").
		SetTextBody("Body")

	resp, err := client.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("Send() error = %v, want nil", err)
	}

	// Check each recipient
	for email, expected := range map[string]struct {
		status int
		msgID  string
	}{
		"recipient1@example.com": {200, "msg-11111"},
		"recipient2@example.com": {200, "msg-22222"},
		"recipient3@example.com": {550, "msg-33333"},
	} {
		status, ok := resp.GetStatus(email)
		if !ok {
			t.Errorf("Expected to find status for %s", email)
			continue
		}
		if status != expected.status {
			t.Errorf("Status for %s = %d, want %d", email, status, expected.status)
		}

		msgID, ok := resp.GetMessageID(email)
		if !ok {
			t.Errorf("Expected to find message ID for %s", email)
			continue
		}
		if msgID != expected.msgID {
			t.Errorf("MessageID for %s = %q, want %q", email, msgID, expected.msgID)
		}
	}
}

func TestClient_Send_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not valid json`))
	}))
	defer server.Close()

	client := NewClient("user", "pass", WithBaseURL(server.URL))

	msg := NewMessage().
		SetSender("sender@example.com").
		AddTo("recipient@example.com").
		SetSubject("Test").
		SetTextBody("Body")

	_, err := client.Send(context.Background(), msg)
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}

	if !strings.Contains(err.Error(), "unmarshal") {
		t.Errorf("Error should mention unmarshal, got: %v", err)
	}
}

func TestClient_Send_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("user", "pass", WithBaseURL(server.URL))

	msg := NewMessage().
		SetSender("sender@example.com").
		AddTo("recipient@example.com").
		SetSubject("Test").
		SetTextBody("Body")

	resp, err := client.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("Send() error = %v, want nil", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}

	if len(resp.Recipients) != 0 {
		t.Errorf("Recipients length = %d, want 0", len(resp.Recipients))
	}
}

func TestClient_Send_WithAttachments(t *testing.T) {
	var receivedMsg Message

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedMsg)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"recipient@example.com": [200, "msg-12345"]}`))
	}))
	defer server.Close()

	client := NewClient("user", "pass", WithBaseURL(server.URL))

	msg := NewMessage().
		SetSender("sender@example.com").
		AddTo("recipient@example.com").
		SetSubject("Test").
		SetTextBody("Body").
		AttachFile("test.txt", "text/plain", []byte("test content"))

	_, err := client.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("Send() error = %v, want nil", err)
	}

	// Verify attachment was sent
	if len(receivedMsg.Attachments) != 1 {
		t.Fatalf("Attachments length = %d, want 1", len(receivedMsg.Attachments))
	}

	if receivedMsg.Attachments[0].Filename != "test.txt" {
		t.Errorf("Filename = %q, want %q", receivedMsg.Attachments[0].Filename, "test.txt")
	}
}
