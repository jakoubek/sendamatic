package sendamatic

import (
	"encoding/json"
	"fmt"
)

// APIError repräsentiert einen API-Fehler
type APIError struct {
	StatusCode       int    `json:"-"`
	Message          string `json:"error"`
	ValidationErrors string `json:"validation_errors,omitempty"`
	JSONPath         string `json:"json_path,omitempty"`
	Sender           string `json:"sender,omitempty"`
	SMTPCode         int    `json:"smtp_code,omitempty"`
}

func (e *APIError) Error() string {
	if e.ValidationErrors != "" {
		return fmt.Sprintf("sendamatic api error (status %d): %s (path: %s)",
			e.StatusCode, e.ValidationErrors, e.JSONPath)
	}
	return fmt.Sprintf("sendamatic api error (status %d): %s", e.StatusCode, e.Message)
}

func parseErrorResponse(statusCode int, body []byte) error {
	var apiErr APIError
	apiErr.StatusCode = statusCode

	if err := json.Unmarshal(body, &apiErr); err != nil {
		// Fallback, falls JSON nicht parsebar ist
		apiErr.Message = string(body)
	}

	return &apiErr
}
