package mailjet

import (
	"testing"
)

func TestNewMailjetConfig(t *testing.T) {
	cfg := NewMailjetConfig("pub-key", "priv-key", "sender@example.com", "Sender Name")

	if cfg.APIKeyPublic != "pub-key" {
		t.Errorf("APIKeyPublic = %q, want %q", cfg.APIKeyPublic, "pub-key")
	}
	if cfg.APIKeyPrivate != "priv-key" {
		t.Errorf("APIKeyPrivate = %q, want %q", cfg.APIKeyPrivate, "priv-key")
	}
	if cfg.OriginAddress != "sender@example.com" {
		t.Errorf("OriginAddress = %q, want %q", cfg.OriginAddress, "sender@example.com")
	}
	if cfg.OriginName != "Sender Name" {
		t.Errorf("OriginName = %q, want %q", cfg.OriginName, "Sender Name")
	}
}
