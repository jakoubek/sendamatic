package sendamatic

import (
	"encoding/json"
	"fmt"
)

// APIError represents an error response from the Sendamatic API.
// It includes the HTTP status code, error message, and optional additional context
// such as validation errors, JSON path information, and SMTP codes.
type APIError struct {
	StatusCode       int    `json:"-"`
	Message          string `json:"error"`
	ValidationErrors string `json:"validation_errors,omitempty"`
	JSONPath         string `json:"json_path,omitempty"`
	Sender           string `json:"sender,omitempty"`
	SMTPCode         int    `json:"smtp_code,omitempty"`
}

// Error implements the error interface and returns a formatted error message.
// If validation errors are present, they are included with the JSON path context.
func (e *APIError) Error() string {
	if e.ValidationErrors != "" {
		return fmt.Sprintf("sendamatic api error (status %d): %s (path: %s)",
			e.StatusCode, e.ValidationErrors, e.JSONPath)
	}
	return fmt.Sprintf("sendamatic api error (status %d): %s", e.StatusCode, e.Message)
}

// parseErrorResponse attempts to parse an API error response body into an APIError.
// If the body cannot be parsed as JSON, it uses the raw body as the error message.
func parseErrorResponse(statusCode int, body []byte) error {
	var apiErr APIError
	apiErr.StatusCode = statusCode

	if err := json.Unmarshal(body, &apiErr); err != nil {
		// Fallback, falls JSON nicht parsebar ist
		apiErr.Message = string(body)
	}

	return &apiErr
}
