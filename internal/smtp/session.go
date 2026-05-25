package smtp

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"

	gosmtp "github.com/emersion/go-smtp"
	"github.com/jo-hoe/go-mail-service/internal/config"
	"github.com/jo-hoe/go-mail-service/internal/mail"
)

// SMTPSession holds per-connection envelope state for one SMTP transaction.
type SMTPSession struct {
	mailService mail.MailService
	auth        config.SMTPAuthConfig
	from        string
	recipients  []string
}

func newSMTPSession(svc mail.MailService, auth config.SMTPAuthConfig) *SMTPSession {
	return &SMTPSession{
		mailService: svc,
		auth:        auth,
	}
}

// AuthPlain validates AUTH PLAIN/LOGIN credentials.
func (s *SMTPSession) AuthPlain(username, password string) error {
	if username != s.auth.Username || password != s.auth.Password {
		return errors.New("invalid credentials")
	}
	return nil
}

// Mail records the envelope sender.
func (s *SMTPSession) Mail(from string, _ *gosmtp.MailOptions) error {
	s.from = from
	return nil
}

// Rcpt appends a recipient to the envelope.
func (s *SMTPSession) Rcpt(to string, _ *gosmtp.RcptOptions) error {
	s.recipients = append(s.recipients, to)
	return nil
}

// Data reads and parses the message, then dispatches it to the mail service.
func (s *SMTPSession) Data(r io.Reader) error {
	if len(s.recipients) == 0 {
		return errors.New("smtp: no recipients")
	}

	parsed, err := parseMessage(r)
	if err != nil {
		slog.Error("smtp: failed to parse message", "error", err)
		return err
	}

	attrs := mail.MailAttributes{
		To:          strings.Join(s.recipients, ","),
		Subject:     parsed.subject,
		HtmlContent: parsed.body,
		From:        s.from,
	}

	if err := s.mailService.SendMail(context.Background(), attrs); err != nil {
		slog.Error("smtp: mail service failed", "error", err)
		return err
	}

	slog.Info("smtp: mail dispatched", "to", attrs.To)
	return nil
}

// Reset clears envelope state. Auth state lives on the connection (managed by go-smtp).
func (s *SMTPSession) Reset() {
	s.from = ""
	s.recipients = nil
}

// Logout is called when the client issues QUIT.
func (s *SMTPSession) Logout() error {
	return nil
}
