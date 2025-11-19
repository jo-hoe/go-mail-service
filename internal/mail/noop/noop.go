package noop

import (
	"context"
	"log"

	"github.com/jo-hoe/go-mail-service/internal/mail"
)

type NoopService struct{}

func NewNoopService() *NoopService {
	return &NoopService{}
}

func (service *NoopService) SendMail(ctx context.Context, attributes mail.MailAttributes) error {
	log.Printf("noop: preparing to send mail - to: %s, subject: %s", attributes.To, attributes.Subject)
	log.Printf("noop: mail details - from: %s (%s), html content length: %d bytes",
		attributes.From, attributes.FromName, len(attributes.HtmlContent))
	log.Printf("noop: mail processed (no actual sending - noop mode)")
	return nil
}
