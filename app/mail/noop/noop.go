package noop

import (
	"context"

	"github.com/jo-hoe/go-mail-service/app/mail"
)

type NoopService struct{}

func NewNoopService() *NoopService {
	return &NoopService{}
}

func (service *NoopService) SendMail(ctx context.Context, attributes mail.MailAttributes) error {
	return nil
}
