package mailjet

// MailjetConfig contains all attributes to initialize the Mailjet mail service.
type MailjetConfig struct {
	APIKeyPublic  string
	APIKeyPrivate string
	OriginAddress string
	OriginName    string
}

// NewMailjetConfig creates a MailjetConfig from the provided credentials and sender identity.
func NewMailjetConfig(apiKeyPublic, apiKeyPrivate, originAddress, originName string) *MailjetConfig {
	return &MailjetConfig{
		APIKeyPublic:  apiKeyPublic,
		APIKeyPrivate: apiKeyPrivate,
		OriginAddress: originAddress,
		OriginName:    originName,
	}
}
