package smtp

import (
	"testing"

	"github.com/jo-hoe/go-mail-service/internal/config"
	"github.com/jo-hoe/go-mail-service/internal/mail/noop"
)

func TestNewSMTPBackend_CreatesSession(t *testing.T) {
	svc := noop.NewNoopService()
	auth := config.SMTPAuthConfig{Required: false}

	backend := NewSMTPBackend(svc, auth)
	session, err := backend.NewSession(nil)
	if err != nil {
		t.Fatalf("NewSession() error: %v", err)
	}
	if session == nil {
		t.Fatal("NewSession() returned nil session")
	}
}
