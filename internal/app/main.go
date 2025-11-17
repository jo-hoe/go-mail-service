package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/jo-hoe/go-mail-service/internal/config"
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

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Validator = &validation.GenericValidator{Validator: validator.New()}

	e.POST("/v1/sendmail", sendMailHandler)
	e.GET("/", probeHandler)

	envService := config.NewEnvService()
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
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err = ctx.Validate(mailAttributes); err != nil {
		return err
	}

	envService := config.NewEnvService()
	isSendGridEnabled, err := envService.Get(IS_SENDGRID_ENABLED_ENV_KEY)
	if err != nil {
		return err
	}
	isMailjetEnabled, err := envService.Get(IS_MAILJET_ENABLED_ENV_KEY)
	if err != nil {
		return err
	}
	isNoopEnabled, err := envService.Get(IS_NOOP_ENABLED_ENV_KEY)
	if err != nil {
		return err
	}

	if strings.ToLower(isNoopEnabled) == "true" {
		noopService := noop.NewNoopService()
		if err = noopService.SendMail(ctx.Request().Context(), *mailAttributes); err != nil {
			return err
		}
	} else if strings.ToLower(isMailjetEnabled) == "true" {
		err = sendWithMailjet(ctx, *mailAttributes)
		if err != nil {
			return err
		}
	} else if strings.ToLower(isSendGridEnabled) == "true" {
		err = sendWithSendGrid(ctx, *mailAttributes)
		if err != nil {
			return err
		}
	}

	return ctx.JSON(http.StatusOK, mailAttributes)
}

func sendWithMailjet(ctx echo.Context, mailAttributes mail.MailAttributes) (err error) {
	mailjetConfig, err := mailjet.NewMailjetConfig(mailAttributes)
	if err != nil {
		return err
	}
	mailService := mailjet.NewMailjetService(mailjetConfig)

	if err = mailService.SendMail(ctx.Request().Context(), mailAttributes); err != nil {
		return err
	}

	return nil
}

func sendWithSendGrid(ctx echo.Context, mailAttributes mail.MailAttributes) (err error) {
	sendgridConfig, err := sendgrid.NewSendGridConfig(mailAttributes)
	if err != nil {
		return err
	}
	mailService := sendgrid.NewSendGridService(sendgridConfig)

	if err = mailService.SendMail(ctx.Request().Context(), mailAttributes); err != nil {
		return err
	}

	return nil
}

func probeHandler(ctx echo.Context) (err error) {
	return ctx.NoContent(http.StatusOK)
}
