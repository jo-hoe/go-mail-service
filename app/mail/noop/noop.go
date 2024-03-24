package noop

import (
	"context"
	"log"

	"github.com/jo-hoe/go-mail-service/app/mail"
)

type NoopService struct{}

func NewNoopService() *NoopService {
	return &NoopService{}
}

func (service *NoopService) SendMail(ctx context.Context, attributes mail.MailAttributes) error {
	log.Printf("Noop service received mail request: %+v", attributes)
	return nil
}
