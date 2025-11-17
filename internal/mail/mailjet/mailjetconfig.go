package mailjet

import (
	"github.com/jo-hoe/go-mail-service/internal/config"
	"github.com/jo-hoe/go-mail-service/internal/mail"
)

// MailjetConfig contains all attributes to initialize the Mailjet mail service
type MailjetConfig struct {
	APIKeyPublic  string
	APIKeyPrivate string
	OriginAddress string
	OriginName    string
}

const apiKeyPublicEnvKey = "MAILJET_API_KEY_PUBLIC"
const apiKeyPrivateEnvKey = "MAILJET_API_KEY_PRIVATE"
const defaultAddressEnvKey = "DEFAULT_FROM_ADDRESS"
const defaultNameEnvKey = "DEFAULT_FROM_NAME"

func NewMailjetConfig(mailAttributes mail.MailAttributes) (result *MailjetConfig, err error) {
	return createConfig(mailAttributes)
}

func createConfig(mailAttributes mail.MailAttributes) (*MailjetConfig, error) {
	envService := config.NewEnvService()

	apiKeyPublic, err := envService.Get(apiKeyPublicEnvKey)
	if err != nil {
		return nil, err
	}

	apiKeyPrivate, err := envService.Get(apiKeyPrivateEnvKey)
	if err != nil {
		return nil, err
	}

	fromAddress, err := getField(mailAttributes.From, defaultAddressEnvKey)
	if err != nil {
		return nil, err
	}

	fromName, err := getField(mailAttributes.FromName, defaultNameEnvKey)
	if err != nil {
		return nil, err
	}

	return &MailjetConfig{
		APIKeyPublic:  apiKeyPublic,
		APIKeyPrivate: apiKeyPrivate,
		OriginAddress: fromAddress,
		OriginName:    fromName,
	}, nil
}

func getField(userInput string, defaultEnvKey string) (result string, err error) {
	envService := config.NewEnvService()
	fromAddress := ""
	if userInput != "" {
		fromAddress = userInput
	} else {
		fromAddress, err = envService.Get(defaultEnvKey)
		if err != nil {
			return "", err
		}
	}
	return fromAddress, nil
}
