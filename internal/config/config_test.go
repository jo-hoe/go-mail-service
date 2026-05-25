package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0600); err != nil {
		t.Fatalf("writeFile %s: %v", p, err)
	}
	return p
}

// yamlPath converts a file path to forward slashes for safe embedding in YAML strings.
func yamlPath(p string) string {
	return strings.ReplaceAll(p, `\`, `/`)
}

func validConfigYAML(smtpAuthRequired bool, smtpPasswordFile, mailjetPubFile, mailjetPrivFile, sendgridKeyFile string) string {
	authRequired := "false"
	passwordFileLine := ""
	if smtpAuthRequired {
		authRequired = "true"
		passwordFileLine = `    passwordFile: "` + yamlPath(smtpPasswordFile) + `"`
	}
	return `logLevel: "info"
sender:
  address: "noreply@example.com"
  name: "Test Service"
http:
  port: 8080
smtp:
  port: 587
  domain: "mail.example.com"
  auth:
    required: ` + authRequired + `
    username: "smtp-user"
` + passwordFileLine + `
  tls:
    enabled: false
provider:
  mailjet:
    enabled: false
    apiKeyPublicFile: "` + yamlPath(mailjetPubFile) + `"
    apiKeyPrivateFile: "` + yamlPath(mailjetPrivFile) + `"
  sendgrid:
    enabled: false
    apiKeyFile: "` + yamlPath(sendgridKeyFile) + `"
  noop:
    enabled: true
`
}

func TestLoad_ValidNoAuth(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeFile(t, dir, "config.yaml", validConfigYAML(false, "", "", "", ""))

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.HTTP.Port != 8080 {
		t.Errorf("http.port = %d, want 8080", cfg.HTTP.Port)
	}
	if cfg.SMTP.Port != 587 {
		t.Errorf("smtp.port = %d, want 587", cfg.SMTP.Port)
	}
	if cfg.Provider.Noop.Enabled != true {
		t.Error("noop.enabled should be true")
	}
}

func TestLoad_WithSMTPAuth(t *testing.T) {
	dir := t.TempDir()
	pwFile := writeFile(t, dir, "password", "secret123\n")
	cfgPath := writeFile(t, dir, "config.yaml", validConfigYAML(true, pwFile, "", "", ""))

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.SMTP.Auth.Password != "secret123" {
		t.Errorf("SMTP password = %q, want %q", cfg.SMTP.Auth.Password, "secret123")
	}
}

func TestLoad_MissingSMTPPasswordFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeFile(t, dir, "config.yaml", validConfigYAML(true, "/nonexistent/password", "", "", ""))

	_, err := Load(cfgPath)
	if err == nil {
		t.Fatal("Load() expected error for missing secret file, got nil")
	}
}

func TestLoad_EmptySMTPPasswordFile(t *testing.T) {
	dir := t.TempDir()
	pwFile := writeFile(t, dir, "password", "   \n")
	cfgPath := writeFile(t, dir, "config.yaml", validConfigYAML(true, pwFile, "", "", ""))

	_, err := Load(cfgPath)
	if err == nil {
		t.Fatal("Load() expected error for empty password, got nil")
	}
}

func TestLoad_MailjetSecrets(t *testing.T) {
	dir := t.TempDir()
	pubFile := writeFile(t, dir, "pub", "pub-key")
	privFile := writeFile(t, dir, "priv", "priv-key")
	yaml := `logLevel: "info"
sender:
  address: "noreply@example.com"
http:
  port: 8080
smtp:
  port: 587
  domain: "mail.example.com"
  auth:
    required: false
  tls:
    enabled: false
provider:
  mailjet:
    enabled: true
    apiKeyPublicFile: "` + yamlPath(pubFile) + `"
    apiKeyPrivateFile: "` + yamlPath(privFile) + `"
  sendgrid:
    enabled: false
  noop:
    enabled: false
`
	cfgPath := writeFile(t, dir, "config.yaml", yaml)

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.Provider.Mailjet.APIKeyPublic != "pub-key" {
		t.Errorf("APIKeyPublic = %q, want %q", cfg.Provider.Mailjet.APIKeyPublic, "pub-key")
	}
	if cfg.Provider.Mailjet.APIKeyPrivate != "priv-key" {
		t.Errorf("APIKeyPrivate = %q, want %q", cfg.Provider.Mailjet.APIKeyPrivate, "priv-key")
	}
}

func TestLoad_MissingConfigFile(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("Load() expected error for missing config file, got nil")
	}
}

func TestValidate_SamePortsRejected(t *testing.T) {
	cfg := &Config{
		Sender: SenderConfig{Address: "a@b.com"},
		HTTP:   HTTPConfig{Port: 8080},
		SMTP: SMTPConfig{
			Port:   8080,
			Domain: "example.com",
		},
		Provider: ProviderConfig{Noop: NoopProviderConfig{Enabled: true}},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() expected error when http.port == smtp.port")
	}
}

func TestValidate_MissingSenderAddress(t *testing.T) {
	cfg := &Config{
		HTTP:     HTTPConfig{Port: 8080},
		SMTP:     SMTPConfig{Port: 587, Domain: "example.com"},
		Provider: ProviderConfig{Noop: NoopProviderConfig{Enabled: true}},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() expected error when sender.address is empty")
	}
}

func TestReadSecretFile_TrimsWhitespace(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "secret", "  mykey\n  ")

	got, err := readSecretFile(p)
	if err != nil {
		t.Fatalf("readSecretFile() error: %v", err)
	}
	if got != "mykey" {
		t.Errorf("readSecretFile() = %q, want %q", got, "mykey")
	}
}

func TestReadSecretFile_EmptyPath(t *testing.T) {
	_, err := readSecretFile("")
	if err == nil {
		t.Fatal("readSecretFile() expected error for empty path")
	}
}
