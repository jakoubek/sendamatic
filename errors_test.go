package sendamatic

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		apiErr   *APIError
		wantText string
	}{
		{
			name: "simple error",
			apiErr: &APIError{
				StatusCode: 400,
				Message:    "Invalid request",
			},
			wantText: "sendamatic api error (status 400): Invalid request",
		},
		{
			name: "error with validation details",
			apiErr: &APIError{
				StatusCode:       422,
				Message:          "Validation failed",
				ValidationErrors: "sender is required",
				JSONPath:         "$.sender",
			},
			wantText: "sendamatic api error (status 422): sender is required (path: $.sender)",
		},
		{
			name: "error with SMTP code",
			apiErr: &APIError{
				StatusCode: 500,
				Message:    "SMTP error",
				SMTPCode:   550,
				Sender:     "test@example.com",
			},
			wantText: "sendamatic api error (status 500): SMTP error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.apiErr.Error()
			if got != tt.wantText {
				t.Errorf("Error() = %q, want %q", got, tt.wantText)
			}
		})
	}
}

func TestParseErrorResponse_ValidJSON(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		want       *APIError
	}{
		{
			name:       "simple error",
			statusCode: 400,
			body:       `{"error": "Invalid API key"}`,
			want: &APIError{
				StatusCode: 400,
				Message:    "Invalid API key",
			},
		},
		{
			name:       "validation error",
			statusCode: 422,
			body:       `{"error": "Validation failed", "validation_errors": "sender is required", "json_path": "$.sender"}`,
			want: &APIError{
				StatusCode:       422,
				Message:          "Validation failed",
				ValidationErrors: "sender is required",
				JSONPath:         "$.sender",
			},
		},
		{
			name:       "smtp error",
			statusCode: 500,
			body:       `{"error": "SMTP error", "smtp_code": 550, "sender": "test@example.com"}`,
			want: &APIError{
				StatusCode: 500,
				Message:    "SMTP error",
				SMTPCode:   550,
				Sender:     "test@example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseErrorResponse(tt.statusCode, []byte(tt.body))

			apiErr, ok := err.(*APIError)
			if !ok {
				t.Fatalf("parseErrorResponse returned %T, want *APIError", err)
			}

			if apiErr.StatusCode != tt.want.StatusCode {
				t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.want.StatusCode)
			}

			if apiErr.Message != tt.want.Message {
				t.Errorf("Message = %q, want %q", apiErr.Message, tt.want.Message)
			}

			if apiErr.ValidationErrors != tt.want.ValidationErrors {
				t.Errorf("ValidationErrors = %q, want %q", apiErr.ValidationErrors, tt.want.ValidationErrors)
			}

			if apiErr.JSONPath != tt.want.JSONPath {
				t.Errorf("JSONPath = %q, want %q", apiErr.JSONPath, tt.want.JSONPath)
			}

			if apiErr.SMTPCode != tt.want.SMTPCode {
				t.Errorf("SMTPCode = %d, want %d", apiErr.SMTPCode, tt.want.SMTPCode)
			}

			if apiErr.Sender != tt.want.Sender {
				t.Errorf("Sender = %q, want %q", apiErr.Sender, tt.want.Sender)
			}
		})
	}
}

func TestParseErrorResponse_InvalidJSON(t *testing.T) {
	statusCode := 500
	body := []byte("Internal Server Error - not JSON")

	err := parseErrorResponse(statusCode, body)

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("parseErrorResponse returned %T, want *APIError", err)
	}

	if apiErr.StatusCode != statusCode {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, statusCode)
	}

	// When JSON parsing fails, the raw body should be used as the message
	if apiErr.Message != string(body) {
		t.Errorf("Message = %q, want %q", apiErr.Message, string(body))
	}
}

func TestParseErrorResponse_EmptyBody(t *testing.T) {
	statusCode := 404
	body := []byte("")

	err := parseErrorResponse(statusCode, body)

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("parseErrorResponse returned %T, want *APIError", err)
	}

	if apiErr.StatusCode != statusCode {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, statusCode)
	}
}

func TestParseErrorResponse_MalformedJSON(t *testing.T) {
	statusCode := 400
	body := []byte(`{"error": "Missing closing brace"`)

	err := parseErrorResponse(statusCode, body)

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("parseErrorResponse returned %T, want *APIError", err)
	}

	if apiErr.StatusCode != statusCode {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, statusCode)
	}

	// Should fall back to raw body as message
	if apiErr.Message != string(body) {
		t.Errorf("Message = %q, want %q", apiErr.Message, string(body))
	}
}

func TestAPIError_JSONRoundtrip(t *testing.T) {
	original := &APIError{
		StatusCode:       422,
		Message:          "Validation error",
		ValidationErrors: "sender format invalid",
		JSONPath:         "$.sender",
		Sender:           "invalid@",
		SMTPCode:         0,
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded APIError
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// StatusCode should not be in JSON (json:"-" tag)
	if strings.Contains(string(data), "StatusCode") {
		t.Error("StatusCode should not be marshaled to JSON")
	}

	// Compare fields (except StatusCode which has json:"-")
	if decoded.Message != original.Message {
		t.Errorf("Message = %q, want %q", decoded.Message, original.Message)
	}

	if decoded.ValidationErrors != original.ValidationErrors {
		t.Errorf("ValidationErrors = %q, want %q", decoded.ValidationErrors, original.ValidationErrors)
	}

	if decoded.JSONPath != original.JSONPath {
		t.Errorf("JSONPath = %q, want %q", decoded.JSONPath, original.JSONPath)
	}
}
