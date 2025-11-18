package sendamatic

// SendResponse represents the response from a send email request.
// It contains the overall HTTP status code and per-recipient delivery information
// including individual status codes and message IDs.
type SendResponse struct {
	StatusCode int
	Recipients map[string][2]interface{} // Email address -> [status code, message ID]
}

// IsSuccess returns true if the email send request was successful (HTTP 200).
// Note that this checks the overall request status; individual recipients
// may still have failed. Use GetStatus to check per-recipient delivery status.
func (r *SendResponse) IsSuccess() bool {
	return r.StatusCode == 200
}

// GetMessageID returns the message ID for a specific recipient email address.
// The message ID can be used to track the email in logs or with the email provider.
// Returns the message ID and true if found, or empty string and false if not found.
func (r *SendResponse) GetMessageID(email string) (string, bool) {
	if info, ok := r.Recipients[email]; ok && len(info) >= 2 {
		if msgID, ok := info[1].(string); ok {
			return msgID, true
		}
	}
	return "", false
}

// GetStatus returns the delivery status code for a specific recipient email address.
// The status code indicates whether the email was accepted for delivery to that recipient.
// Returns the status code and true if found, or 0 and false if not found.
//
// Note: The API returns status codes as JSON numbers which are decoded as float64,
// so this method performs the necessary type conversion to int.
func (r *SendResponse) GetStatus(email string) (int, bool) {
	if info, ok := r.Recipients[email]; ok && len(info) >= 1 {
		if status, ok := info[0].(float64); ok {
			return int(status), true
		}
	}
	return 0, false
}
