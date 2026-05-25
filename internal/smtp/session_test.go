package smtp

import (
	"context"
	"strings"
	"testing"

	gosmtp "github.com/emersion/go-smtp"
	"github.com/jo-hoe/go-mail-service/internal/config"
	"github.com/jo-hoe/go-mail-service/internal/mail"
)

// captureService records the last MailAttributes passed to SendMail.
type captureService struct {
	last mail.MailAttributes
}

func (c *captureService) SendMail(_ context.Context, attrs mail.MailAttributes) error {
	c.last = attrs
	return nil
}

func newTestSession(authRequired bool, username, password string) (*SMTPSession, *captureService) {
	svc := &captureService{}
	auth := config.SMTPAuthConfig{
		Required: authRequired,
		Username: username,
		Password: password,
	}
	return newSMTPSession(svc, auth), svc
}

func TestSMTPSession_AuthPlain_Valid(t *testing.T) {
	s, _ := newTestSession(true, "user", "pass")
	if err := s.AuthPlain("user", "pass"); err != nil {
		t.Errorf("AuthPlain() unexpected error: %v", err)
	}
}

func TestSMTPSession_AuthPlain_Invalid(t *testing.T) {
	s, _ := newTestSession(true, "user", "pass")
	if err := s.AuthPlain("user", "wrong"); err == nil {
		t.Error("AuthPlain() expected error for wrong password, got nil")
	}
}

func TestSMTPSession_Mail_SetsFrom(t *testing.T) {
	s, _ := newTestSession(false, "", "")
	if err := s.Mail("sender@example.com", &gosmtp.MailOptions{}); err != nil {
		t.Fatalf("Mail() error: %v", err)
	}
	if s.from != "sender@example.com" {
		t.Errorf("from = %q, want %q", s.from, "sender@example.com")
	}
}

func TestSMTPSession_Rcpt_AppendsRecipient(t *testing.T) {
	s, _ := newTestSession(false, "", "")
	_ = s.Rcpt("a@example.com", &gosmtp.RcptOptions{})
	_ = s.Rcpt("b@example.com", &gosmtp.RcptOptions{})
	if len(s.recipients) != 2 {
		t.Errorf("recipients len = %d, want 2", len(s.recipients))
	}
}

func TestSMTPSession_Reset_ClearsEnvelope(t *testing.T) {
	s, _ := newTestSession(false, "", "")
	s.from = "x@example.com"
	s.recipients = []string{"y@example.com"}
	s.Reset()
	if s.from != "" {
		t.Errorf("from after Reset = %q, want empty", s.from)
	}
	if len(s.recipients) != 0 {
		t.Errorf("recipients after Reset = %v, want empty", s.recipients)
	}
}

func TestSMTPSession_Data_DispatchesMail(t *testing.T) {
	s, svc := newTestSession(false, "", "")
	_ = s.Mail("sender@example.com", &gosmtp.MailOptions{})
	_ = s.Rcpt("to@example.com", &gosmtp.RcptOptions{})

	raw := "Subject: Hello\r\nContent-Type: text/plain\r\n\r\nTest body"
	if err := s.Data(strings.NewReader(raw)); err != nil {
		t.Fatalf("Data() error: %v", err)
	}

	if svc.last.Subject != "Hello" {
		t.Errorf("subject = %q, want %q", svc.last.Subject, "Hello")
	}
	if svc.last.To != "to@example.com" {
		t.Errorf("to = %q, want %q", svc.last.To, "to@example.com")
	}
	if svc.last.From != "sender@example.com" {
		t.Errorf("from = %q, want %q", svc.last.From, "sender@example.com")
	}
}

func TestSMTPSession_Data_MultipleRecipients(t *testing.T) {
	s, svc := newTestSession(false, "", "")
	_ = s.Mail("sender@example.com", &gosmtp.MailOptions{})
	_ = s.Rcpt("a@example.com", &gosmtp.RcptOptions{})
	_ = s.Rcpt("b@example.com", &gosmtp.RcptOptions{})

	raw := "Subject: Multi\r\nContent-Type: text/plain\r\n\r\nBody"
	if err := s.Data(strings.NewReader(raw)); err != nil {
		t.Fatalf("Data() error: %v", err)
	}

	if svc.last.To != "a@example.com,b@example.com" {
		t.Errorf("to = %q, want comma-joined recipients", svc.last.To)
	}
}

func TestSMTPSession_Data_NoRecipients(t *testing.T) {
	s, _ := newTestSession(false, "", "")
	raw := "Subject: X\r\n\r\nBody"
	if err := s.Data(strings.NewReader(raw)); err == nil {
		t.Error("Data() expected error when no recipients, got nil")
	}
}

func TestSMTPSession_Logout(t *testing.T) {
	s, _ := newTestSession(false, "", "")
	if err := s.Logout(); err != nil {
		t.Errorf("Logout() error: %v", err)
	}
}
