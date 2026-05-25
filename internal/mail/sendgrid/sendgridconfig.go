package sendgrid

// SendGridConfig contains all attributes to initialize the SendGrid mail service.
type SendGridConfig struct {
	APIKey        string
	OriginAddress string
	OriginName    string
}

// NewSendGridConfig creates a SendGridConfig from the provided credentials and sender identity.
func NewSendGridConfig(apiKey, originAddress, originName string) *SendGridConfig {
	return &SendGridConfig{
		APIKey:        apiKey,
		OriginAddress: originAddress,
		OriginName:    originName,
	}
}
