package smtp

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"

	gosmtp "github.com/emersion/go-smtp"
	"github.com/jo-hoe/go-mail-service/internal/config"
	"github.com/jo-hoe/go-mail-service/internal/mail"
)

const maxMessageBytes = 8 * 1024 * 1024 // 8 MB

// SMTPServer wraps a go-smtp server and provides lifecycle methods.
type SMTPServer struct {
	server *gosmtp.Server
}

// NewSMTPServer creates an SMTPServer configured from cfg, using svc for mail dispatch.
func NewSMTPServer(cfg *config.SMTPConfig, svc mail.MailService) (*SMTPServer, error) {
	backend := NewSMTPBackend(svc, cfg.Auth)

	s := gosmtp.NewServer(backend)
	s.Domain = cfg.Domain
	s.Addr = fmt.Sprintf(":%d", cfg.Port)
	s.MaxMessageBytes = maxMessageBytes
	s.AllowInsecureAuth = !cfg.Auth.Required

	if cfg.TLS.Enabled {
		tlsCfg, err := loadTLS(cfg.TLS)
		if err != nil {
			return nil, fmt.Errorf("smtp: loading TLS config: %w", err)
		}
		s.TLSConfig = tlsCfg
	}

	return &SMTPServer{server: s}, nil
}

// Start begins listening and blocks until the server stops.
func (s *SMTPServer) Start() error {
	slog.Info("smtp: server starting", "addr", s.server.Addr)
	if s.server.TLSConfig != nil {
		return s.server.ListenAndServeTLS()
	}
	return s.server.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *SMTPServer) Shutdown(ctx context.Context) error {
	slog.Info("smtp: shutting down")
	return s.server.Shutdown(ctx)
}

func loadTLS(cfg config.SMTPTLSConfig) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}
