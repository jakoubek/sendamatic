package sendamatic

import (
	"encoding/base64"
	"errors"
	"os"
)

// Message represents an email message with all its components including recipients,
// content, headers, and attachments. Messages are constructed using the fluent builder
// pattern provided by the setter methods.
type Message struct {
	To          []string     `json:"to"`
	CC          []string     `json:"cc,omitempty"`
	BCC         []string     `json:"bcc,omitempty"`
	Sender      string       `json:"sender"`
	Subject     string       `json:"subject"`
	TextBody    string       `json:"text_body,omitempty"`
	HTMLBody    string       `json:"html_body,omitempty"`
	Headers     []Header     `json:"headers,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Header represents a custom email header as a name-value pair.
type Header struct {
	Header string `json:"header"`
	Value  string `json:"value"`
}

// Attachment represents an email attachment with its filename, MIME type, and base64-encoded data.
type Attachment struct {
	Filename string `json:"filename"`
	Data     string `json:"data"` // Base64-encoded file content
	MimeType string `json:"mimetype"`
}

// NewMessage creates and returns a new empty Message with initialized slices for recipients,
// headers, and attachments. Use the setter methods to populate the message fields.
func NewMessage() *Message {
	return &Message{
		To:          []string{},
		CC:          []string{},
		BCC:         []string{},
		Headers:     []Header{},
		Attachments: []Attachment{},
	}
}

// AddTo adds a recipient email address to the To field.
// Returns the message for method chaining.
func (m *Message) AddTo(email string) *Message {
	m.To = append(m.To, email)
	return m
}

// AddCC adds a recipient email address to the CC (carbon copy) field.
// Returns the message for method chaining.
func (m *Message) AddCC(email string) *Message {
	m.CC = append(m.CC, email)
	return m
}

// AddBCC adds a recipient email address to the BCC (blind carbon copy) field.
// Returns the message for method chaining.
func (m *Message) AddBCC(email string) *Message {
	m.BCC = append(m.BCC, email)
	return m
}

// SetSender sets the sender email address for the message.
// Returns the message for method chaining.
func (m *Message) SetSender(email string) *Message {
	m.Sender = email
	return m
}

// SetSubject sets the email subject line.
// Returns the message for method chaining.
func (m *Message) SetSubject(subject string) *Message {
	m.Subject = subject
	return m
}

// SetTextBody sets the plain text body of the email.
// Returns the message for method chaining.
func (m *Message) SetTextBody(body string) *Message {
	m.TextBody = body
	return m
}

// SetHTMLBody sets the HTML body of the email.
// Returns the message for method chaining.
func (m *Message) SetHTMLBody(body string) *Message {
	m.HTMLBody = body
	return m
}

// AddHeader adds a custom email header with the specified name and value.
// Common examples include "Reply-To", "X-Priority", or custom application headers.
// Returns the message for method chaining.
func (m *Message) AddHeader(name, value string) *Message {
	m.Headers = append(m.Headers, Header{
		Header: name,
		Value:  value,
	})
	return m
}

// AttachFile adds a file attachment to the message from a byte slice.
// The data is automatically base64-encoded for transmission.
// Returns the message for method chaining.
func (m *Message) AttachFile(filename, mimeType string, data []byte) *Message {
	m.Attachments = append(m.Attachments, Attachment{
		Filename: filename,
		Data:     base64.StdEncoding.EncodeToString(data),
		MimeType: mimeType,
	})
	return m
}

// AttachFileFromPath reads a file from the filesystem and adds it as an attachment.
// The filename is extracted from the path. Returns an error if the file cannot be read.
// The file data is automatically base64-encoded for transmission.
func (m *Message) AttachFileFromPath(path, mimeType string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Extrahiere Dateinamen aus Pfad
	filename := path
	if idx := len(path) - 1; idx >= 0 {
		for i := idx; i >= 0; i-- {
			if path[i] == '/' || path[i] == '\\' {
				filename = path[i+1:]
				break
			}
		}
	}

	m.AttachFile(filename, mimeType, data)
	return nil
}

// Validate checks whether the message meets all required criteria for sending.
// It returns an error if any validation rules are violated:
//   - At least one recipient is required
//   - Maximum of 255 recipients allowed
//   - Sender must be specified
//   - Subject must be specified
//   - Either TextBody or HTMLBody (or both) must be provided
func (m *Message) Validate() error {
	if len(m.To) == 0 {
		return errors.New("at least one recipient required")
	}
	if len(m.To) > 255 {
		return errors.New("maximum 255 recipients allowed")
	}
	if m.Sender == "" {
		return errors.New("sender is required")
	}
	if m.Subject == "" {
		return errors.New("subject is required")
	}
	if m.TextBody == "" && m.HTMLBody == "" {
		return errors.New("either text_body or html_body is required")
	}
	return nil
}
