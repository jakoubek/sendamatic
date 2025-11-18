package sendamatic

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestNewMessage(t *testing.T) {
	msg := NewMessage()

	if msg == nil {
		t.Fatal("NewMessage returned nil")
	}

	if msg.To == nil || msg.CC == nil || msg.BCC == nil {
		t.Error("Slices not initialized")
	}

	if msg.Headers == nil || msg.Attachments == nil {
		t.Error("Headers or Attachments not initialized")
	}
}

func TestMessageBuilderMethods(t *testing.T) {
	msg := NewMessage().
		SetSender("sender@example.com").
		AddTo("to@example.com").
		AddCC("cc@example.com").
		AddBCC("bcc@example.com").
		SetSubject("Test Subject").
		SetTextBody("Test Body").
		SetHTMLBody("<p>Test Body</p>").
		AddHeader("X-Custom", "value")

	if msg.Sender != "sender@example.com" {
		t.Errorf("Sender = %q, want %q", msg.Sender, "sender@example.com")
	}

	if len(msg.To) != 1 || msg.To[0] != "to@example.com" {
		t.Errorf("To = %v, want [to@example.com]", msg.To)
	}

	if len(msg.CC) != 1 || msg.CC[0] != "cc@example.com" {
		t.Errorf("CC = %v, want [cc@example.com]", msg.CC)
	}

	if len(msg.BCC) != 1 || msg.BCC[0] != "bcc@example.com" {
		t.Errorf("BCC = %v, want [bcc@example.com]", msg.BCC)
	}

	if msg.Subject != "Test Subject" {
		t.Errorf("Subject = %q, want %q", msg.Subject, "Test Subject")
	}

	if msg.TextBody != "Test Body" {
		t.Errorf("TextBody = %q, want %q", msg.TextBody, "Test Body")
	}

	if msg.HTMLBody != "<p>Test Body</p>" {
		t.Errorf("HTMLBody = %q, want %q", msg.HTMLBody, "<p>Test Body</p>")
	}

	if len(msg.Headers) != 1 {
		t.Fatalf("Headers length = %d, want 1", len(msg.Headers))
	}

	if msg.Headers[0].Header != "X-Custom" || msg.Headers[0].Value != "value" {
		t.Errorf("Header = %+v, want {X-Custom value}", msg.Headers[0])
	}
}

func TestAddMultipleRecipients(t *testing.T) {
	msg := NewMessage().
		AddTo("to1@example.com").
		AddTo("to2@example.com").
		AddTo("to3@example.com")

	if len(msg.To) != 3 {
		t.Errorf("To length = %d, want 3", len(msg.To))
	}

	expected := []string{"to1@example.com", "to2@example.com", "to3@example.com"}
	for i, email := range expected {
		if msg.To[i] != email {
			t.Errorf("To[%d] = %q, want %q", i, msg.To[i], email)
		}
	}
}

func TestAttachFile(t *testing.T) {
	msg := NewMessage()
	data := []byte("test file content")

	msg.AttachFile("test.txt", "text/plain", data)

	if len(msg.Attachments) != 1 {
		t.Fatalf("Attachments length = %d, want 1", len(msg.Attachments))
	}

	att := msg.Attachments[0]
	if att.Filename != "test.txt" {
		t.Errorf("Filename = %q, want %q", att.Filename, "test.txt")
	}

	if att.MimeType != "text/plain" {
		t.Errorf("MimeType = %q, want %q", att.MimeType, "text/plain")
	}

	// Verify base64 encoding
	decoded, err := base64.StdEncoding.DecodeString(att.Data)
	if err != nil {
		t.Fatalf("Failed to decode base64: %v", err)
	}

	if string(decoded) != string(data) {
		t.Errorf("Decoded data = %q, want %q", decoded, data)
	}
}

func TestAttachFileFromPath(t *testing.T) {
	msg := NewMessage()

	testFile := filepath.Join("testdata", "test.txt")
	err := msg.AttachFileFromPath(testFile, "text/plain")
	if err != nil {
		t.Fatalf("AttachFileFromPath failed: %v", err)
	}

	if len(msg.Attachments) != 1 {
		t.Fatalf("Attachments length = %d, want 1", len(msg.Attachments))
	}

	att := msg.Attachments[0]
	if att.Filename != "test.txt" {
		t.Errorf("Filename = %q, want %q", att.Filename, "test.txt")
	}

	// Verify content
	decoded, err := base64.StdEncoding.DecodeString(att.Data)
	if err != nil {
		t.Fatalf("Failed to decode base64: %v", err)
	}

	expected, _ := os.ReadFile(testFile)
	if string(decoded) != string(expected) {
		t.Errorf("File content mismatch")
	}
}

func TestAttachFileFromPath_NonExistent(t *testing.T) {
	msg := NewMessage()

	err := msg.AttachFileFromPath("nonexistent.txt", "text/plain")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestAttachMultipleFiles(t *testing.T) {
	msg := NewMessage().
		AttachFile("file1.txt", "text/plain", []byte("content1")).
		AttachFile("file2.pdf", "application/pdf", []byte("content2"))

	if len(msg.Attachments) != 2 {
		t.Errorf("Attachments length = %d, want 2", len(msg.Attachments))
	}
}

func TestValidate_Success(t *testing.T) {
	tests := []struct {
		name string
		msg  *Message
	}{
		{
			name: "valid with text body",
			msg: NewMessage().
				SetSender("sender@example.com").
				AddTo("to@example.com").
				SetSubject("Subject").
				SetTextBody("Body"),
		},
		{
			name: "valid with html body",
			msg: NewMessage().
				SetSender("sender@example.com").
				AddTo("to@example.com").
				SetSubject("Subject").
				SetHTMLBody("<p>Body</p>"),
		},
		{
			name: "valid with both bodies",
			msg: NewMessage().
				SetSender("sender@example.com").
				AddTo("to@example.com").
				SetSubject("Subject").
				SetTextBody("Body").
				SetHTMLBody("<p>Body</p>"),
		},
		{
			name: "valid with multiple recipients",
			msg: NewMessage().
				SetSender("sender@example.com").
				AddTo("to1@example.com").
				AddTo("to2@example.com").
				AddCC("cc@example.com").
				AddBCC("bcc@example.com").
				SetSubject("Subject").
				SetTextBody("Body"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.Validate()
			if err != nil {
				t.Errorf("Validate() error = %v, want nil", err)
			}
		})
	}
}

func TestValidate_Errors(t *testing.T) {
	tests := []struct {
		name        string
		msg         *Message
		wantErrText string
	}{
		{
			name:        "no recipients",
			msg:         NewMessage().SetSender("sender@example.com").SetSubject("Subject").SetTextBody("Body"),
			wantErrText: "at least one recipient required",
		},
		{
			name:        "no sender",
			msg:         NewMessage().AddTo("to@example.com").SetSubject("Subject").SetTextBody("Body"),
			wantErrText: "sender is required",
		},
		{
			name:        "no subject",
			msg:         NewMessage().SetSender("sender@example.com").AddTo("to@example.com").SetTextBody("Body"),
			wantErrText: "subject is required",
		},
		{
			name:        "no body",
			msg:         NewMessage().SetSender("sender@example.com").AddTo("to@example.com").SetSubject("Subject"),
			wantErrText: "either text_body or html_body is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.Validate()
			if err == nil {
				t.Fatal("Validate() error = nil, want error")
			}
			if err.Error() != tt.wantErrText {
				t.Errorf("Validate() error = %q, want %q", err.Error(), tt.wantErrText)
			}
		})
	}
}

func TestValidate_TooManyRecipients(t *testing.T) {
	msg := NewMessage().
		SetSender("sender@example.com").
		SetSubject("Subject").
		SetTextBody("Body")

	// Add 256 recipients (more than the limit of 255)
	for i := 0; i < 256; i++ {
		msg.AddTo("recipient@example.com")
	}

	err := msg.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want error for too many recipients")
	}

	expected := "maximum 255 recipients allowed"
	if err.Error() != expected {
		t.Errorf("Validate() error = %q, want %q", err.Error(), expected)
	}
}
