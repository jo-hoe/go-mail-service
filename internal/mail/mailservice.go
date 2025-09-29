package mail

import "context"

// Service forwards E-Mail to a set of receivers
type MailService interface {
	SendMail(ctx context.Context, attributes MailAttributes) error
}
