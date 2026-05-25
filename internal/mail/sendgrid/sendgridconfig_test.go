package sendgrid

import (
	"testing"
)

func TestNewSendGridConfig(t *testing.T) {
	cfg := NewSendGridConfig("api-key", "sender@example.com", "Sender Name")

	if cfg.APIKey != "api-key" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "api-key")
	}
	if cfg.OriginAddress != "sender@example.com" {
		t.Errorf("OriginAddress = %q, want %q", cfg.OriginAddress, "sender@example.com")
	}
	if cfg.OriginName != "Sender Name" {
		t.Errorf("OriginName = %q, want %q", cfg.OriginName, "Sender Name")
	}
}
