package sendgrid

import (
	"github.com/jo-hoe/go-mail-service/internal/config"
	"github.com/jo-hoe/go-mail-service/internal/mail"
)

// SendGridConfig contain all attributes to initialize the SendGrid mail service
type SendGridConfig struct {
	APIKey        string
	OriginAddress string
	OriginName    string
}

const apiEnvKey = "SENDGRID_API_KEY"
const defaultAddressEnvKey = "DEFAULT_FROM_ADDRESS"
const defaultNameEnvKey = "DEFAULT_FROM_NAME"

func NewSendGridConfig(mailAttributes mail.MailAttributes) (result *SendGridConfig, err error) {
	return createConfig(mailAttributes)
}

func createConfig(mailAttributes mail.MailAttributes) (*SendGridConfig, error) {
	envService := config.NewEnvService()

	apiKey, err := envService.Get(apiEnvKey)
	if err != nil {
		return nil, err
	}

	fromAddress, err := mail.GetFieldOrDefault(mailAttributes.From, defaultAddressEnvKey)
	if err != nil {
		return nil, err
	}

	fromName, err := mail.GetFieldOrDefault(mailAttributes.FromName, defaultNameEnvKey)
	if err != nil {
		return nil, err
	}

	return &SendGridConfig{
		APIKey:        apiKey,
		OriginAddress: fromAddress,
		OriginName:    fromName,
	}, nil
}
