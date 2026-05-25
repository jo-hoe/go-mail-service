package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator"
	"github.com/jo-hoe/go-mail-service/internal/config"
	"github.com/jo-hoe/go-mail-service/internal/logging"
	"github.com/jo-hoe/go-mail-service/internal/mail"
	"github.com/jo-hoe/go-mail-service/internal/mail/mailjet"
	"github.com/jo-hoe/go-mail-service/internal/mail/noop"
	"github.com/jo-hoe/go-mail-service/internal/mail/sendgrid"
	appsmtp "github.com/jo-hoe/go-mail-service/internal/smtp"
	"github.com/jo-hoe/go-mail-service/internal/validation"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const configPathEnvKey = "CONFIG_PATH"
const defaultConfigPath = "/config/config.yaml"
const shutdownTimeout = 10 * time.Second

func main() {
	cfgPath := os.Getenv(configPathEnvKey)
	if cfgPath == "" {
		cfgPath = defaultConfigPath
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logging.New(logging.Config{
		Level:     logging.ParseLevel(cfg.LogLevel),
		AddSource: true,
		JSON:      false,
	})

	svc, err := resolveMailService(cfg)
	if err != nil {
		slog.Error("failed to resolve mail service", "error", err)
		os.Exit(1)
	}

	e := buildHTTPServer(svc)
	smtpServer, err := appsmtp.NewSMTPServer(&cfg.SMTP, svc)
	if err != nil {
		slog.Error("failed to create smtp server", "error", err)
		os.Exit(1)
	}

	go func() {
		addr := fmt.Sprintf(":%d", cfg.HTTP.Port)
		slog.Info("http: server starting", "addr", addr)
		if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server stopped", "error", err)
		}
	}()

	go func() {
		if err := smtpServer.Start(); err != nil {
			slog.Error("smtp server stopped", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := errors.Join(e.Shutdown(ctx), smtpServer.Shutdown(ctx)); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}

func buildHTTPServer(svc mail.MailService) *echo.Echo {
	e := echo.New()
	e.Use(middleware.RequestLoggerWithConfig(requestLoggerConfig()))
	e.Use(middleware.Recover())
	e.Validator = &validation.GenericValidator{Validator: validator.New()}

	e.POST("/v1/sendmail", sendMailHandler(svc))
	e.GET("/", probeHandler)

	return e
}

func sendMailHandler(svc mail.MailService) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		attrs := new(mail.MailAttributes)
		if err := ctx.Bind(attrs); err != nil {
			slog.Error("failed to bind mail attributes", "error", err)
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err := ctx.Validate(attrs); err != nil {
			slog.Error("failed to validate mail attributes", "error", err)
			return err
		}

		slog.Info("received mail request")
		if err := svc.SendMail(ctx.Request().Context(), *attrs); err != nil {
			slog.Error("failed to send mail", "error", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return ctx.JSON(http.StatusOK, attrs)
	}
}

func probeHandler(ctx echo.Context) error {
	return ctx.NoContent(http.StatusOK)
}

// resolveMailService returns the highest-priority enabled mail provider.
func resolveMailService(cfg *config.Config) (mail.MailService, error) {
	p := cfg.Provider
	switch {
	case p.Mailjet.Enabled:
		mCfg := mailjet.NewMailjetConfig(
			p.Mailjet.APIKeyPublic,
			p.Mailjet.APIKeyPrivate,
			cfg.Sender.Address,
			cfg.Sender.Name,
		)
		return mailjet.NewMailjetService(mCfg), nil
	case p.SendGrid.Enabled:
		sCfg := sendgrid.NewSendGridConfig(
			p.SendGrid.APIKey,
			cfg.Sender.Address,
			cfg.Sender.Name,
		)
		return sendgrid.NewSendGridService(sCfg), nil
	case p.Noop.Enabled:
		return noop.NewNoopService(), nil
	default:
		return nil, fmt.Errorf("no mail provider is enabled")
	}
}

func requestLoggerConfig() middleware.RequestLoggerConfig {
	return middleware.RequestLoggerConfig{
		Skipper:      func(c echo.Context) bool { return c.Request().Method == http.MethodGet && c.Path() == "/" },
		LogStatus:    true,
		LogLatency:   true,
		LogURI:       true,
		LogMethod:    true,
		LogError:     true,
		LogRemoteIP:  true,
		LogUserAgent: true,
		LogValuesFunc: func(_ echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error != nil {
				slog.Error("http request",
					"method", v.Method,
					"uri", v.URI,
					"status", v.Status,
					"latency", v.Latency,
					"remote_ip", v.RemoteIP,
					"user_agent", v.UserAgent,
					"error", v.Error,
				)
			} else {
				slog.Info("http request",
					"method", v.Method,
					"uri", v.URI,
					"status", v.Status,
					"latency", v.Latency,
					"remote_ip", v.RemoteIP,
					"user_agent", v.UserAgent,
				)
			}
			return nil
		},
	}
}
