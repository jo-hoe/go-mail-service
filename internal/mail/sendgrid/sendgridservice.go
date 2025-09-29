package sendgrid

import (
	"context"
	"fmt"
	"strings"

	"github.com/jo-hoe/go-mail-service/internal/mail"

	"github.com/sendgrid/sendgrid-go"
	sgmail "github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendGridService implements MailService
type SendGridService struct {
	config   *SendGridConfig
	messages []*sgmail.SGMailV3
}

// NewSendGridService creates a SendGridService using an already initializes service
func NewSendGridService(config *SendGridConfig) *SendGridService {
	return &SendGridService{
		config:   config,
		messages: make([]*sgmail.SGMailV3, 0),
	}
}

func (service *SendGridService) SendMail(ctx context.Context, attributes mail.MailAttributes) error {
	message := service.createMessage(attributes)
	return service.sendRequest(ctx, message)
}

// createMessages a sendgrid message
func (service *SendGridService) createMessage(attributes mail.MailAttributes) *sgmail.SGMailV3 {
	// create new *SGMailV3
	mailObject := sgmail.NewV3Mail()

	from := sgmail.NewEmail(service.config.OriginName, service.config.OriginAddress)
	content := sgmail.NewContent("text/html", attributes.HtmlContent)

	mailObject.SetFrom(from)
	mailObject.AddContent(content)

	// create new *Personalization
	personalization := sgmail.NewPersonalization()

	personalization.Subject = attributes.Subject
	// populate `personalization` with data
	emails := []*sgmail.Email{}

	mailAddresses := strings.Split(attributes.To, ",")
	for _, mailAddress := range mailAddresses {
		mail, _ := sgmail.ParseEmail(mailAddress)
		emails = append(emails, mail)
	}

	personalization.AddTos(emails...)

	// add `personalization` to `m`
	mailObject.AddPersonalizations(personalization)
	return mailObject
}

func (service *SendGridService) sendRequest(ctx context.Context, mailObject *sgmail.SGMailV3) error {
	request := sendgrid.GetRequest(
		service.config.APIKey,
		"/v3/mail/send",
		"https://api.sendgrid.com",
	)

	request.Method = "POST"
	request.Body = sgmail.GetRequestBody(mailObject)
	result, err := sendgrid.MakeRequestWithContext(ctx, request)

	if result.StatusCode != 202 {
		return fmt.Errorf("SendGrid could not send mail. [%d]: %s", result.StatusCode, result.Body)
	}

	return err
}
