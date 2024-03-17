package sendgrid

import (
	"github.com/jo-hoe/go-mail-service/app/mail"
	"testing"
)

func Test_Init(t *testing.T) {
	config := getTestConfig()

	sender := NewSendGridService(&config)

	if sender == nil {
		t.Errorf("Sendgrid not initialized")
	}
}

func Test_AddMessage(t *testing.T) {
	config := getTestConfig()

	sender := NewSendGridService(&config)
	message := sender.createMessage(mail.MailAttributes{
		To:          []string{"test@test.com"},
		Subject:     "test",
		HtmlContent: "test content",
	})

	if message == nil {
		t.Error("Expected message not to be nil")
	}
}

func getTestConfig() SendGridConfig {
	return SendGridConfig{
		APIKey:        "testkey",
		OriginAddress: "keyaddress",
		OriginName:    "testname",
	}
}
