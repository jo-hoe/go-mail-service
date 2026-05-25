package smtp

import (
	gosmtp "github.com/emersion/go-smtp"
	"github.com/jo-hoe/go-mail-service/internal/config"
	"github.com/jo-hoe/go-mail-service/internal/mail"
)

// SMTPBackend implements the go-smtp Backend interface.
// It creates a new session for each incoming connection.
type SMTPBackend struct {
	mailService mail.MailService
	auth        config.SMTPAuthConfig
}

// NewSMTPBackend creates an SMTPBackend using the provided mail service and auth config.
func NewSMTPBackend(svc mail.MailService, auth config.SMTPAuthConfig) *SMTPBackend {
	return &SMTPBackend{
		mailService: svc,
		auth:        auth,
	}
}

// NewSession creates a fresh session for an incoming SMTP connection.
func (b *SMTPBackend) NewSession(_ *gosmtp.Conn) (gosmtp.Session, error) {
	return newSMTPSession(b.mailService, b.auth), nil
}
