package sendgrid

import (
	"io"
	"os"

	"github.com/jo-hoe/go-mail-service/app/config"
	"github.com/jo-hoe/go-mail-service/app/mail"
)

// SendGridConfig contain all attributes to initialize the SendGrid mail service
type SendGridConfig struct {
	APIKey        string
	OriginAddress string
	OriginName    string
}

const apiKeyFilePath = "./run/secrets/sendgrid_api_key.txt"
const defaultAddressEnvKey = "DEFAULT_FROM_ADDRESS"
const defaultNameEnvKey = "DEFAULT_FROM_NAME"

func NewSendGridConfig(mailAttributes mail.MailAttributes) (result *SendGridConfig, err error) {
	file, err := os.Open(apiKeyFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return createConfig(mailAttributes, file)
}

func createConfig(mailAttributes mail.MailAttributes, reader io.Reader) (*SendGridConfig, error) {
	secretService := config.NewSecretFileService()

	apiKey, err := secretService.Get(reader)
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

	return &SendGridConfig{
		APIKey:        apiKey,
		OriginAddress: fromAddress,
		OriginName:    fromName,
	}, nil
}

func getField(userInput string, defaultEnvKey string) (result string, err error) {
	secretService := config.NewEnvService()
	fromAddress := ""
	if userInput != "" {
		fromAddress = userInput
	} else {
		fromAddress, err = secretService.Get(defaultEnvKey)
		if err != nil {
			return "", err
		}
	}
	return fromAddress, nil
}
