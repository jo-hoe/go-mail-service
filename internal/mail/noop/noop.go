package noop

import (
	"context"
	"log/slog"

	"github.com/jo-hoe/go-mail-service/internal/mail"
)

type NoopService struct{}

func NewNoopService() *NoopService {
	return &NoopService{}
}

func (service *NoopService) SendMail(ctx context.Context, attributes mail.MailAttributes) error {
	slog.Info("noop: preparing to send mail", "to", attributes.To, "subject", attributes.Subject)
	slog.Debug("noop: mail details", "from", attributes.From, "from_name", attributes.FromName, "html_len", len(attributes.HtmlContent))
	slog.Info("noop: mail processed (no actual sending - noop mode)")
	return nil
}
