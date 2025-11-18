package sendamatic

import (
	"encoding/base64"
	"errors"
	"os"
)

// Message repräsentiert eine E-Mail-Nachricht
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

// Header repräsentiert einen benutzerdefinierten E-Mail-Header
type Header struct {
	Header string `json:"header"`
	Value  string `json:"value"`
}

// Attachment repräsentiert einen E-Mail-Anhang
type Attachment struct {
	Filename string `json:"filename"`
	Data     string `json:"data"` // Base64-kodiert
	MimeType string `json:"mimetype"`
}

// NewMessage erstellt eine neue Message
func NewMessage() *Message {
	return &Message{
		To:          []string{},
		CC:          []string{},
		BCC:         []string{},
		Headers:     []Header{},
		Attachments: []Attachment{},
	}
}

// AddTo fügt einen Empfänger hinzu
func (m *Message) AddTo(email string) *Message {
	m.To = append(m.To, email)
	return m
}

// AddCC fügt einen CC-Empfänger hinzu
func (m *Message) AddCC(email string) *Message {
	m.CC = append(m.CC, email)
	return m
}

// AddBCC fügt einen BCC-Empfänger hinzu
func (m *Message) AddBCC(email string) *Message {
	m.BCC = append(m.BCC, email)
	return m
}

// SetSender setzt den Absender
func (m *Message) SetSender(email string) *Message {
	m.Sender = email
	return m
}

// SetSubject setzt den Betreff
func (m *Message) SetSubject(subject string) *Message {
	m.Subject = subject
	return m
}

// SetTextBody setzt den Text-Körper
func (m *Message) SetTextBody(body string) *Message {
	m.TextBody = body
	return m
}

// SetHTMLBody setzt den HTML-Körper
func (m *Message) SetHTMLBody(body string) *Message {
	m.HTMLBody = body
	return m
}

// AddHeader fügt einen benutzerdefinierten Header hinzu
func (m *Message) AddHeader(name, value string) *Message {
	m.Headers = append(m.Headers, Header{
		Header: name,
		Value:  value,
	})
	return m
}

// AttachFile fügt eine Datei als Anhang hinzu
func (m *Message) AttachFile(filename, mimeType string, data []byte) *Message {
	m.Attachments = append(m.Attachments, Attachment{
		Filename: filename,
		Data:     base64.StdEncoding.EncodeToString(data),
		MimeType: mimeType,
	})
	return m
}

// AttachFileFromPath lädt eine Datei vom Dateisystem und fügt sie als Anhang hinzu
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

// Validate prüft, ob die Message gültig ist
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
