package smtp

import (
	"strings"
	"testing"
)

func TestParseMessage_PlainText(t *testing.T) {
	raw := "Subject: Hello\r\nContent-Type: text/plain\r\n\r\nHello world"
	msg, err := parseMessage(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("parseMessage() error: %v", err)
	}
	if msg.subject != "Hello" {
		t.Errorf("subject = %q, want %q", msg.subject, "Hello")
	}
	if !strings.Contains(msg.body, "Hello world") {
		t.Errorf("body = %q, expected to contain %q", msg.body, "Hello world")
	}
	if !strings.HasPrefix(msg.body, "<pre>") {
		t.Errorf("plain text should be wrapped in <pre>, got: %q", msg.body)
	}
}

func TestParseMessage_HTML(t *testing.T) {
	raw := "Subject: Test\r\nContent-Type: text/html\r\n\r\n<p>Hello</p>"
	msg, err := parseMessage(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("parseMessage() error: %v", err)
	}
	if msg.subject != "Test" {
		t.Errorf("subject = %q, want %q", msg.subject, "Test")
	}
	if msg.body != "<p>Hello</p>" {
		t.Errorf("body = %q, want %q", msg.body, "<p>Hello</p>")
	}
}

func TestParseMessage_MultipartPrefersHTML(t *testing.T) {
	raw := "Subject: Multi\r\nContent-Type: multipart/alternative; boundary=\"boundary\"\r\n\r\n" +
		"--boundary\r\n" +
		"Content-Type: text/plain\r\n\r\n" +
		"Plain text\r\n" +
		"--boundary\r\n" +
		"Content-Type: text/html\r\n\r\n" +
		"<p>HTML</p>\r\n" +
		"--boundary--\r\n"

	msg, err := parseMessage(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("parseMessage() error: %v", err)
	}
	if msg.body != "<p>HTML</p>" {
		t.Errorf("body = %q, expected HTML part", msg.body)
	}
}

func TestParseMessage_MultipartPlainFallback(t *testing.T) {
	raw := "Subject: Plain\r\nContent-Type: multipart/alternative; boundary=\"boundary\"\r\n\r\n" +
		"--boundary\r\n" +
		"Content-Type: text/plain\r\n\r\n" +
		"Only plain\r\n" +
		"--boundary--\r\n"

	msg, err := parseMessage(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("parseMessage() error: %v", err)
	}
	if !strings.HasPrefix(msg.body, "<pre>") {
		t.Errorf("expected plain fallback wrapped in <pre>, got: %q", msg.body)
	}
	if !strings.Contains(msg.body, "Only plain") {
		t.Errorf("body = %q, expected to contain %q", msg.body, "Only plain")
	}
}

func TestParseMessage_NoContentType(t *testing.T) {
	raw := "Subject: No CT\r\n\r\nHello without content type"
	msg, err := parseMessage(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("parseMessage() error: %v", err)
	}
	if !strings.Contains(msg.body, "Hello without content type") {
		t.Errorf("body = %q, expected to contain raw text", msg.body)
	}
}
