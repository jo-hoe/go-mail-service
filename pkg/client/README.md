# Go Mail Service Client

A simple Go client library for interacting with the Go Mail Service API.

## Installation

```bash
go get github.com/jo-hoe/go-mail-service/pkg/client
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/jo-hoe/go-mail-service/pkg/client"
)

func main() {
    // Create a new client
    mailClient := client.NewClient("http://localhost:80")

    // Create a mail request
    request := client.MailRequest{
        To:          "recipient@example.com",
        Subject:     "Hello from Go Mail Service",
        HtmlContent: "<h1>Hello!</h1><p>This is a test email.</p>",
        From:        "sender@example.com",
        FromName:    "Test Sender",
    }

    // Send the email
    response, err := mailClient.SendMail(context.Background(), request)
    if err != nil {
        log.Fatalf("Failed to send email: %v", err)
    }

    fmt.Printf("Email sent successfully to: %s\n", response.To)
}
```

## API Reference

### Creating a Client

#### `NewClient(baseURL string, options ...ClientOption) *Client`

Creates a new mail service client with the given base URL and optional configuration.

```go
// Basic client
client := client.NewClient("http://localhost:80")

// Client with custom timeout
client := client.NewClient("http://localhost:80", client.WithTimeout(10*time.Second))

// Client with custom HTTP client
httpClient := &http.Client{Timeout: 15 * time.Second}
client := client.NewClient("http://localhost:80", client.WithHTTPClient(httpClient))
```

### Client Options

#### `WithTimeout(timeout time.Duration) ClientOption`

Sets a custom timeout for HTTP requests (default: 30 seconds).

#### `WithHTTPClient(httpClient *http.Client) ClientOption`

Sets a custom HTTP client for making requests.

### Sending Mail

#### `SendMail(ctx context.Context, request MailRequest) (*MailResponse, error)`

Sends an email using the mail service.

**MailRequest fields:**
- `To` (required): Recipient email address(es), comma-separated for multiple recipients
- `Subject` (required): Email subject line
- `HtmlContent` (required): Email body content in HTML format
- `From` (optional): Sender email address
- `FromName` (optional): Display name for the sender

```go
request := client.MailRequest{
    To:          "user@example.com",
    Subject:     "Test Email",
    HtmlContent: "<p>Hello World!</p>",
    From:        "sender@company.com",
    FromName:    "Company Name",
}

response, err := mailClient.SendMail(context.Background(), request)
if err != nil {
    // Handle error
    if errorResp, ok := err.(client.ErrorResponse); ok {
        fmt.Printf("HTTP Error %d: %s\n", errorResp.Code, errorResp.Message)
    } else {
        fmt.Printf("Error: %v\n", err)
    }
    return
}

fmt.Printf("Email sent to: %s\n", response.To)
```

### Health Check

#### `HealthCheck(ctx context.Context) error`

Performs a health check against the mail service.

```go
err := mailClient.HealthCheck(context.Background())
if err != nil {
    log.Printf("Mail service is not healthy: %v", err)
} else {
    log.Println("Mail service is healthy!")
}
```

## Advanced Usage

### Multiple Recipients

Send emails to multiple recipients by providing a comma-separated list of email addresses:

```go
request := client.MailRequest{
    To:          "user1@example.com,user2@example.com,user3@example.com",
    Subject:     "Team Notification",
    HtmlContent: "<h2>Team Update</h2><p>This is for the whole team.</p>",
}
```

### Context with Timeout

Use context for request timeouts and cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

response, err := mailClient.SendMail(ctx, request)
```

### Error Handling

The client provides structured error handling:

```go
response, err := mailClient.SendMail(context.Background(), request)
if err != nil {
    switch e := err.(type) {
    case client.ErrorResponse:
        // HTTP error from the service
        fmt.Printf("Service error (HTTP %d): %s\n", e.Code, e.Message)
    default:
        // Other errors (network, validation, etc.)
        fmt.Printf("Client error: %v\n", err)
    }
    return
}
```

## Error Types

### `ErrorResponse`

Represents an HTTP error response from the mail service:

```go
type ErrorResponse struct {
    Message string `json:"message"`
    Code    int    `json:"code"`
}
```

Common HTTP status codes:
- `400 Bad Request`: Invalid request data (missing required fields, invalid email format, etc.)
- `500 Internal Server Error`: Server-side error

## Examples

The code examples above demonstrate the main usage patterns for the client library.

## Configuration

The mail service itself can be configured with environment variables:
- `API_PORT`: Port number for the service (default: 80)
- `IS_NOOP_ENABLED`: Enable no-op mode for testing (default: true)
- `IS_SENDGRID_ENABLED`: Enable SendGrid integration (default: false)

## Testing

Run the client tests:

```bash
go test ./pkg/client/...
```

Run with verbose output:

```bash
go test -v ./pkg/client/...
```

## License

This client library is part of the Go Mail Service project and follows the same license terms.
