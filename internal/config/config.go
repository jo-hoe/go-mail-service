package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is the complete application configuration loaded from a YAML file.
type Config struct {
	LogLevel string         `yaml:"logLevel"`
	Sender   SenderConfig   `yaml:"sender"`
	HTTP     HTTPConfig     `yaml:"http"`
	SMTP     SMTPConfig     `yaml:"smtp"`
	Provider ProviderConfig `yaml:"provider"`
}

// SenderConfig holds the default outbound sender identity.
type SenderConfig struct {
	Address string `yaml:"address"`
	Name    string `yaml:"name"`
}

// HTTPConfig holds HTTP server settings.
type HTTPConfig struct {
	Port int `yaml:"port"`
}

// SMTPConfig holds SMTP server settings.
type SMTPConfig struct {
	Port   int            `yaml:"port"`
	Domain string         `yaml:"domain"`
	Auth   SMTPAuthConfig `yaml:"auth"`
	TLS    SMTPTLSConfig  `yaml:"tls"`
}

// SMTPAuthConfig holds SMTP authentication settings.
// Password is resolved from PasswordFile at load time and stored in Password.
type SMTPAuthConfig struct {
	Required     bool   `yaml:"required"`
	Username     string `yaml:"username"`
	PasswordFile string `yaml:"passwordFile"`
	Password     string `yaml:"-"` // resolved at load time, never in YAML output
}

// SMTPTLSConfig holds optional TLS settings for the SMTP server.
type SMTPTLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"certFile"`
	KeyFile  string `yaml:"keyFile"`
}

// ProviderConfig selects and configures the active mail provider.
type ProviderConfig struct {
	Mailjet  MailjetProviderConfig  `yaml:"mailjet"`
	SendGrid SendGridProviderConfig `yaml:"sendgrid"`
	Noop     NoopProviderConfig     `yaml:"noop"`
}

// MailjetProviderConfig holds Mailjet settings.
// Credentials are resolved from the file paths at load time.
type MailjetProviderConfig struct {
	Enabled          bool   `yaml:"enabled"`
	APIKeyPublicFile string `yaml:"apiKeyPublicFile"`
	APIKeyPrivateFile string `yaml:"apiKeyPrivateFile"`
	APIKeyPublic     string `yaml:"-"` // resolved at load time
	APIKeyPrivate    string `yaml:"-"` // resolved at load time
}

// SendGridProviderConfig holds SendGrid settings.
// Credentials are resolved from the file path at load time.
type SendGridProviderConfig struct {
	Enabled    bool   `yaml:"enabled"`
	APIKeyFile string `yaml:"apiKeyFile"`
	APIKey     string `yaml:"-"` // resolved at load time
}

// NoopProviderConfig enables the no-op provider for development.
type NoopProviderConfig struct {
	Enabled bool `yaml:"enabled"`
}

// Load reads the YAML file at path, resolves all secret files, and validates the result.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- config path is operator-supplied (CONFIG_PATH), not user input
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if err := cfg.resolveSecrets(); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// resolveSecrets reads all referenced secret files and populates the in-memory credential fields.
func (c *Config) resolveSecrets() error {
	if c.SMTP.Auth.Required {
		pw, err := readSecretFile(c.SMTP.Auth.PasswordFile)
		if err != nil {
			return fmt.Errorf("smtp auth password: %w", err)
		}
		c.SMTP.Auth.Password = pw
	}

	if c.Provider.Mailjet.Enabled {
		pub, err := readSecretFile(c.Provider.Mailjet.APIKeyPublicFile)
		if err != nil {
			return fmt.Errorf("mailjet apiKeyPublic: %w", err)
		}
		priv, err := readSecretFile(c.Provider.Mailjet.APIKeyPrivateFile)
		if err != nil {
			return fmt.Errorf("mailjet apiKeyPrivate: %w", err)
		}
		c.Provider.Mailjet.APIKeyPublic = pub
		c.Provider.Mailjet.APIKeyPrivate = priv
	}

	if c.Provider.SendGrid.Enabled {
		key, err := readSecretFile(c.Provider.SendGrid.APIKeyFile)
		if err != nil {
			return fmt.Errorf("sendgrid apiKey: %w", err)
		}
		c.Provider.SendGrid.APIKey = key
	}

	return nil
}

// Validate checks that all required fields are present and consistent.
func (c *Config) Validate() error {
	var errs []error

	if c.Sender.Address == "" {
		errs = append(errs, errors.New("sender.address is required"))
	}
	if c.HTTP.Port <= 0 {
		errs = append(errs, errors.New("http.port must be greater than 0"))
	}
	if c.SMTP.Port <= 0 {
		errs = append(errs, errors.New("smtp.port must be greater than 0"))
	}
	if c.HTTP.Port == c.SMTP.Port {
		errs = append(errs, fmt.Errorf("http.port and smtp.port must be different (both are %d)", c.HTTP.Port))
	}
	if c.SMTP.Domain == "" {
		errs = append(errs, errors.New("smtp.domain is required"))
	}

	if c.SMTP.Auth.Required {
		if c.SMTP.Auth.Username == "" {
			errs = append(errs, errors.New("smtp.auth.username is required when auth is required"))
		}
		if c.SMTP.Auth.Password == "" {
			errs = append(errs, errors.New("smtp.auth.password resolved to empty — check passwordFile"))
		}
	}

	if c.SMTP.TLS.Enabled {
		if c.SMTP.TLS.CertFile == "" {
			errs = append(errs, errors.New("smtp.tls.certFile is required when TLS is enabled"))
		}
		if c.SMTP.TLS.KeyFile == "" {
			errs = append(errs, errors.New("smtp.tls.keyFile is required when TLS is enabled"))
		}
	}

	if c.Provider.Mailjet.Enabled {
		if c.Provider.Mailjet.APIKeyPublic == "" {
			errs = append(errs, errors.New("mailjet apiKeyPublic resolved to empty"))
		}
		if c.Provider.Mailjet.APIKeyPrivate == "" {
			errs = append(errs, errors.New("mailjet apiKeyPrivate resolved to empty"))
		}
	}

	if c.Provider.SendGrid.Enabled && c.Provider.SendGrid.APIKey == "" {
		errs = append(errs, errors.New("sendgrid apiKey resolved to empty"))
	}

	warnMultipleProviders(c)

	if !c.Provider.Mailjet.Enabled && !c.Provider.SendGrid.Enabled && !c.Provider.Noop.Enabled {
		slog.Warn("no mail provider is enabled — mail will not be sent")
	}

	return errors.Join(errs...)
}

func warnMultipleProviders(c *Config) {
	enabled := 0
	if c.Provider.Mailjet.Enabled {
		enabled++
	}
	if c.Provider.SendGrid.Enabled {
		enabled++
	}
	if c.Provider.Noop.Enabled {
		enabled++
	}
	if enabled > 1 {
		slog.Warn("multiple providers enabled — only highest priority will be used (mailjet > sendgrid > noop)")
	}
}

// readSecretFile reads a single-line secret from a file, trimming whitespace.
func readSecretFile(path string) (string, error) {
	if path == "" {
		return "", errors.New("secret file path is empty")
	}
	data, err := os.ReadFile(path) // #nosec G304 -- secret file paths come from operator-supplied config, not user input
	if err != nil {
		return "", fmt.Errorf("reading secret file %q: %w", path, err)
	}
	return strings.TrimSpace(string(data)), nil
}
