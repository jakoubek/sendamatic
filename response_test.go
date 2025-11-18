package sendamatic

import (
	"encoding/json"
	"testing"
)

func TestSendResponse_IsSuccess(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"success 200", 200, true},
		{"bad request 400", 400, false},
		{"unauthorized 401", 401, false},
		{"server error 500", 500, false},
		{"created 201", 201, false}, // Only 200 is considered success
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &SendResponse{StatusCode: tt.statusCode}
			got := resp.IsSuccess()
			if got != tt.want {
				t.Errorf("IsSuccess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendResponse_GetMessageID(t *testing.T) {
	tests := []struct {
		name       string
		recipients map[string][2]interface{}
		email      string
		wantID     string
		wantOK     bool
	}{
		{
			name: "existing recipient",
			recipients: map[string][2]interface{}{
				"test@example.com": {float64(200), "msg-12345"},
			},
			email:  "test@example.com",
			wantID: "msg-12345",
			wantOK: true,
		},
		{
			name: "non-existent recipient",
			recipients: map[string][2]interface{}{
				"test@example.com": {float64(200), "msg-12345"},
			},
			email:  "other@example.com",
			wantID: "",
			wantOK: false,
		},
		{
			name: "multiple recipients",
			recipients: map[string][2]interface{}{
				"test1@example.com": {float64(200), "msg-11111"},
				"test2@example.com": {float64(200), "msg-22222"},
				"test3@example.com": {float64(400), "msg-33333"},
			},
			email:  "test2@example.com",
			wantID: "msg-22222",
			wantOK: true,
		},
		{
			name:       "empty recipients",
			recipients: map[string][2]interface{}{},
			email:      "test@example.com",
			wantID:     "",
			wantOK:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &SendResponse{
				StatusCode: 200,
				Recipients: tt.recipients,
			}

			gotID, gotOK := resp.GetMessageID(tt.email)
			if gotID != tt.wantID {
				t.Errorf("GetMessageID() id = %q, want %q", gotID, tt.wantID)
			}
			if gotOK != tt.wantOK {
				t.Errorf("GetMessageID() ok = %v, want %v", gotOK, tt.wantOK)
			}
		})
	}
}

func TestSendResponse_GetStatus(t *testing.T) {
	tests := []struct {
		name       string
		recipients map[string][2]interface{}
		email      string
		wantStatus int
		wantOK     bool
	}{
		{
			name: "existing recipient with success",
			recipients: map[string][2]interface{}{
				"test@example.com": {float64(200), "msg-12345"},
			},
			email:      "test@example.com",
			wantStatus: 200,
			wantOK:     true,
		},
		{
			name: "existing recipient with error",
			recipients: map[string][2]interface{}{
				"test@example.com": {float64(400), "msg-12345"},
			},
			email:      "test@example.com",
			wantStatus: 400,
			wantOK:     true,
		},
		{
			name: "non-existent recipient",
			recipients: map[string][2]interface{}{
				"test@example.com": {float64(200), "msg-12345"},
			},
			email:      "other@example.com",
			wantStatus: 0,
			wantOK:     false,
		},
		{
			name: "multiple recipients",
			recipients: map[string][2]interface{}{
				"test1@example.com": {float64(200), "msg-11111"},
				"test2@example.com": {float64(550), "msg-22222"},
				"test3@example.com": {float64(200), "msg-33333"},
			},
			email:      "test2@example.com",
			wantStatus: 550,
			wantOK:     true,
		},
		{
			name:       "empty recipients",
			recipients: map[string][2]interface{}{},
			email:      "test@example.com",
			wantStatus: 0,
			wantOK:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &SendResponse{
				StatusCode: 200,
				Recipients: tt.recipients,
			}

			gotStatus, gotOK := resp.GetStatus(tt.email)
			if gotStatus != tt.wantStatus {
				t.Errorf("GetStatus() status = %d, want %d", gotStatus, tt.wantStatus)
			}
			if gotOK != tt.wantOK {
				t.Errorf("GetStatus() ok = %v, want %v", gotOK, tt.wantOK)
			}
		})
	}
}

func TestSendResponse_JSONUnmarshal(t *testing.T) {
	// Test that we can properly unmarshal the API response format
	jsonResp := `{
		"test1@example.com": [200, "msg-11111"],
		"test2@example.com": [400, "msg-22222"]
	}`

	var recipients map[string][2]interface{}
	err := json.Unmarshal([]byte(jsonResp), &recipients)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	resp := &SendResponse{
		StatusCode: 200,
		Recipients: recipients,
	}

	// Test first recipient
	status1, ok1 := resp.GetStatus("test1@example.com")
	if !ok1 {
		t.Error("Expected to find test1@example.com")
	}
	if status1 != 200 {
		t.Errorf("Status for test1 = %d, want 200", status1)
	}

	msgID1, ok1 := resp.GetMessageID("test1@example.com")
	if !ok1 {
		t.Error("Expected to find message ID for test1@example.com")
	}
	if msgID1 != "msg-11111" {
		t.Errorf("MessageID for test1 = %q, want %q", msgID1, "msg-11111")
	}

	// Test second recipient
	status2, ok2 := resp.GetStatus("test2@example.com")
	if !ok2 {
		t.Error("Expected to find test2@example.com")
	}
	if status2 != 400 {
		t.Errorf("Status for test2 = %d, want 400", status2)
	}

	msgID2, ok2 := resp.GetMessageID("test2@example.com")
	if !ok2 {
		t.Error("Expected to find message ID for test2@example.com")
	}
	if msgID2 != "msg-22222" {
		t.Errorf("MessageID for test2 = %q, want %q", msgID2, "msg-22222")
	}
}

func TestSendResponse_GetStatus_Float64Conversion(t *testing.T) {
	// Explicitly test the float64 to int conversion
	// This mimics how JSON unmarshaling works with numbers
	resp := &SendResponse{
		StatusCode: 200,
		Recipients: map[string][2]interface{}{
			"test@example.com": {float64(200.0), "msg-12345"},
		},
	}

	status, ok := resp.GetStatus("test@example.com")
	if !ok {
		t.Fatal("Expected to find recipient")
	}

	if status != 200 {
		t.Errorf("GetStatus() = %d, want 200", status)
	}
}

func TestSendResponse_GetMessageID_InvalidType(t *testing.T) {
	// Test behavior when message ID is not a string
	resp := &SendResponse{
		StatusCode: 200,
		Recipients: map[string][2]interface{}{
			"test@example.com": {float64(200), 12345}, // number instead of string
		},
	}

	msgID, ok := resp.GetMessageID("test@example.com")
	if ok {
		t.Error("Expected ok = false when message ID is not a string")
	}
	if msgID != "" {
		t.Errorf("Expected empty string, got %q", msgID)
	}
}

func TestSendResponse_GetStatus_InvalidType(t *testing.T) {
	// Test behavior when status is not a number
	resp := &SendResponse{
		StatusCode: 200,
		Recipients: map[string][2]interface{}{
			"test@example.com": {"OK", "msg-12345"}, // string instead of number
		},
	}

	status, ok := resp.GetStatus("test@example.com")
	if ok {
		t.Error("Expected ok = false when status is not a number")
	}
	if status != 0 {
		t.Errorf("Expected 0, got %d", status)
	}
}
