package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jo-hoe/go-mail-service/internal/config"
	"github.com/jo-hoe/go-mail-service/internal/logging"
	"github.com/jo-hoe/go-mail-service/internal/mail"
	"github.com/jo-hoe/go-mail-service/internal/mail/mailjet"
	"github.com/jo-hoe/go-mail-service/internal/mail/noop"
	"github.com/jo-hoe/go-mail-service/internal/mail/sendgrid"
	"github.com/jo-hoe/go-mail-service/internal/validation"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const API_PORT_ENV_KEY = "API_PORT"
const IS_NOOP_ENABLED_ENV_KEY = "IS_NOOP_ENABLED"
const IS_SENDGRID_ENABLED_ENV_KEY = "IS_SENDGRID_ENABLED"
const IS_MAILJET_ENABLED_ENV_KEY = "IS_MAILJET_ENABLED"
const LOG_LEVEL_ENV_KEY = "LOG_LEVEL"

func shouldSkipRequestLog(c echo.Context) bool {
	// Do not log root probe endpoint
	return c.Request().Method == http.MethodGet && c.Path() == "/"
}

func main() {
	e := echo.New()

	envService := config.NewEnvService()
	// initialize slog with level from env (defaults to info)
	levelStr, _ := envService.Get(LOG_LEVEL_ENV_KEY)
	logging.New(logging.Config{
		Level:     logging.ParseLevel(levelStr),
		AddSource: true,
		JSON:      false,
	})

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		Skipper:      shouldSkipRequestLog,
		LogStatus:    true,
		LogLatency:   true,
		LogURI:       true,
		LogMethod:    true,
		LogError:     true,
		LogRemoteIP:  true,
		LogUserAgent: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
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
	}))
	e.Use(middleware.Recover())
	e.Validator = &validation.GenericValidator{Validator: validator.New()}

	e.POST("/v1/sendmail", sendMailHandler)
	e.GET("/", probeHandler)

	port, err := envService.Get(API_PORT_ENV_KEY)
	if err != nil {
		e.Logger.Fatal(err)
	}

	// start server
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}

func sendMailHandler(ctx echo.Context) (err error) {
	mailAttributes := new(mail.MailAttributes)
	if err = ctx.Bind(mailAttributes); err != nil {
		slog.Error("failed to bind mail attributes", "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err = ctx.Validate(mailAttributes); err != nil {
		slog.Error("failed to validate mail attributes", "error", err)
		return err
	}

	slog.Info("received mail request")

	// Get provider flags
	providerFlags, err := getProviderFlags()
	if err != nil {
		slog.Error("failed to get provider flags", "error", err)
		return err
	}

	// Check for multiple enabled providers and warn
	checkMultipleProviders(ctx, providerFlags)

	// Send mail using the highest priority enabled provider
	if providerFlags.isMailjetEnabled {
		return sendMailWithProvider(ctx, *mailAttributes, "Mailjet", "highest", sendWithMailjet)
	} else if providerFlags.isSendGridEnabled {
		return sendMailWithProvider(ctx, *mailAttributes, "SendGrid", "medium", sendWithSendGrid)
	} else if providerFlags.isNoopEnabled {
		return sendMailWithProvider(ctx, *mailAttributes, "Noop", "lowest - development/testing only", func(ctx echo.Context, attrs mail.MailAttributes) error {
			noopService := noop.NewNoopService()
			return noopService.SendMail(ctx.Request().Context(), attrs)
		})
	}

	slog.Warn("no mail provider is enabled - mail not sent")
	return ctx.JSON(http.StatusOK, mailAttributes)
}

type providerFlags struct {
	isMailjetEnabled  bool
	isSendGridEnabled bool
	isNoopEnabled     bool
}

func getProviderFlags() (*providerFlags, error) {
	envService := config.NewEnvService()

	isMailjetEnabled, err := getEnvAsBool(envService, IS_MAILJET_ENABLED_ENV_KEY)
	if err != nil {
		return nil, err
	}

	isSendGridEnabled, err := getEnvAsBool(envService, IS_SENDGRID_ENABLED_ENV_KEY)
	if err != nil {
		return nil, err
	}

	isNoopEnabled, err := getEnvAsBool(envService, IS_NOOP_ENABLED_ENV_KEY)
	if err != nil {
		return nil, err
	}

	return &providerFlags{
		isMailjetEnabled:  isMailjetEnabled,
		isSendGridEnabled: isSendGridEnabled,
		isNoopEnabled:     isNoopEnabled,
	}, nil
}

func getEnvAsBool(envService *config.EnvService, key string) (bool, error) {
	value, err := envService.Get(key)
	if err != nil {
		return false, err
	}
	return strings.ToLower(value) == "true", nil
}

func checkMultipleProviders(ctx echo.Context, flags *providerFlags) {
	enabledCount := 0
	if flags.isMailjetEnabled {
		enabledCount++
	}
	if flags.isSendGridEnabled {
		enabledCount++
	}
	if flags.isNoopEnabled {
		enabledCount++
	}

	if enabledCount > 1 {
		slog.Warn("multiple mail providers are enabled - only one will be used based on priority: Mailjet → SendGrid → Noop", "enabled_count", enabledCount)
	}
}

type mailSender func(echo.Context, mail.MailAttributes) error

func sendMailWithProvider(ctx echo.Context, mailAttributes mail.MailAttributes, providerName string, priority string, sender mailSender) error {
	slog.Info("using provider", "provider", providerName, "priority", priority)

	if err := sender(ctx, mailAttributes); err != nil {
		slog.Error("provider failed", "provider", strings.ToLower(providerName), "error", err)
		return err
	}

	if providerName == "Noop" {
		slog.Info("mail processed", "provider", providerName)
	} else {
		slog.Info("mail sent", "provider", providerName)
	}

	return ctx.JSON(http.StatusOK, mailAttributes)
}

func sendWithMailjet(ctx echo.Context, mailAttributes mail.MailAttributes) error {
	mailjetConfig, err := mailjet.NewMailjetConfig(mailAttributes)
	if err != nil {
		return err
	}
	mailService := mailjet.NewMailjetService(mailjetConfig)
	return mailService.SendMail(ctx.Request().Context(), mailAttributes)
}

func sendWithSendGrid(ctx echo.Context, mailAttributes mail.MailAttributes) error {
	sendgridConfig, err := sendgrid.NewSendGridConfig(mailAttributes)
	if err != nil {
		return err
	}
	mailService := sendgrid.NewSendGridService(sendgridConfig)
	return mailService.SendMail(ctx.Request().Context(), mailAttributes)
}

func probeHandler(ctx echo.Context) (err error) {
	return ctx.NoContent(http.StatusOK)
}
