# Sendamatic Go Client

[![Mirror on GitHub](https://img.shields.io/badge/mirror-GitHub-blue)](https://github.com/jakoubek/sendamatic)
[![Go Reference](https://pkg.go.dev/badge/code.beautifulmachines.dev/jakoubek/sendamatic.svg)](https://pkg.go.dev/code.beautifulmachines.dev/jakoubek/sendamatic)
[![Go Report Card](https://goreportcard.com/badge/code.beautifulmachines.dev/jakoubek/sendamatic)](https://goreportcard.com/report/code.beautifulmachines.dev/jakoubek/sendamatic)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A Go client library for the [Sendamatic](https://www.sendamatic.net) email delivery API.

## Features

- Simple and idiomatic Go API
- Context support for timeouts and cancellation
- Fluent message builder interface
- Support for HTML and plain text emails
- File attachments with automatic base64 encoding
- Custom headers
- Multiple recipients (To, CC, BCC)
- Comprehensive error handling

## Installation
```bash
go get code.beautifulmachines.dev/jakoubek/sendamatic
```

## Quick Start
```go
package main

import (
    "context"
    "log"

    "code.beautifulmachines.dev/jakoubek/sendamatic"
)

func main() {
    // Create client
    client := sendamatic.NewClient("your-user-id", "your-password")

    // Build message
    msg := sendamatic.NewMessage().
        SetSender("sender@example.com").
        AddTo("recipient@example.com").
        SetSubject("Hello from Sendamatic").
        SetTextBody("This is a test message.").
        SetHTMLBody("<h1>Hello!</h1><p>This is a test message.</p>")

    // Send email
    resp, err := client.Send(context.Background(), msg)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Email sent successfully: %d", resp.StatusCode)
}
```

## Usage Examples

### Basic Email
```go
msg := sendamatic.NewMessage().
    SetSender("sender@example.com").
    AddTo("recipient@example.com").
    SetSubject("Hello World").
    SetTextBody("This is a plain text email.")

resp, err := client.Send(context.Background(), msg)
```

### HTML Email with Multiple Recipients
```go
msg := sendamatic.NewMessage().
    SetSender("newsletter@example.com").
    AddTo("user1@example.com").
    AddTo("user2@example.com").
    AddCC("manager@example.com").
    AddBCC("archive@example.com").
    SetSubject("Monthly Newsletter").
    SetHTMLBody("<h1>Newsletter</h1><p>Your monthly update...</p>").
    SetTextBody("Newsletter - Your monthly update...")
```

### Email with Attachments
```go
// From file path
msg := sendamatic.NewMessage().
    SetSender("sender@example.com").
    AddTo("recipient@example.com").
    SetSubject("Invoice").
    SetTextBody("Please find your invoice attached.")

err := msg.AttachFileFromPath("./invoice.pdf", "application/pdf")
if err != nil {
    log.Fatal(err)
}

// Or from byte slice
pdfData := []byte{...}
msg.AttachFile("invoice.pdf", "application/pdf", pdfData)
```

### Custom Headers
```go
msg := sendamatic.NewMessage().
    SetSender("sender@example.com").
    AddTo("recipient@example.com").
    SetSubject("Custom Headers").
    SetTextBody("Email with custom headers").
    AddHeader("Reply-To", "support@example.com").
    AddHeader("X-Priority", "1")
```

### With Timeout
```go
client := sendamatic.NewClient(
    "user-id",
    "password",
    sendamatic.WithTimeout(45*time.Second),
)

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := client.Send(ctx, msg)
```

### Custom HTTP Client
```go
httpClient := &http.Client{
    Timeout: 60 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns: 10,
    },
}

client := sendamatic.NewClient(
    "user-id",
    "password",
    sendamatic.WithHTTPClient(httpClient),
)
```

## Configuration Options

The client supports various configuration options via the functional options pattern:
```go
client := sendamatic.NewClient(
    "user-id",
    "password",
    sendamatic.WithBaseURL("https://custom.api.url"),
    sendamatic.WithTimeout(60*time.Second),
    sendamatic.WithHTTPClient(customHTTPClient),
)
```

## Response Handling

The `SendResponse` provides methods to check the delivery status:
```go
resp, err := client.Send(ctx, msg)
if err != nil {
    log.Fatal(err)
}

// Check overall success
if resp.IsSuccess() {
    log.Println("Email sent successfully")
}

// Check individual recipient status
for email := range resp.Recipients {
    if status, ok := resp.GetStatus(email); ok {
        log.Printf("Recipient %s: status %d", email, status)
    }
    if msgID, ok := resp.GetMessageID(email); ok {
        log.Printf("Message ID: %s", msgID)
    }
}
```

## Error Handling

The library provides typed errors for better error handling:
```go
resp, err := client.Send(ctx, msg)
if err != nil {
    var apiErr *sendamatic.APIError
    if errors.As(err, &apiErr) {
        log.Printf("API error (status %d): %s", apiErr.StatusCode, apiErr.Message)
        if apiErr.ValidationErrors != "" {
            log.Printf("Validation: %s", apiErr.ValidationErrors)
        }
    } else {
        log.Printf("Other error: %v", err)
    }
}
```

## Requirements

- Go 1.22 or higher
- Valid Sendamatic account with API credentials

## API Credentials

Your API credentials consist of:
- **User ID**: Your Mail Credential User ID
- **Password**: Your Mail Credential Password

Find these in your Sendamatic dashboard under Mail Credentials.

## Documentation

For detailed API documentation, visit:
- [Sendamatic API Documentation](https://docs.sendamatic.net/api/send/)
- [Sendamatic Website](https://www.sendamatic.net)

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## Author

Oliver Jakoubek ([info@jakoubek.net](mailto:info@jakoubek.net))