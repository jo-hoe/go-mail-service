package main

import (
	"fmt"
	"net/http"

	"github.com/jo-hoe/go-mail-service/app/mail"
	"github.com/jo-hoe/go-mail-service/app/mail/sendgrid"
	"github.com/jo-hoe/go-mail-service/app/secret"
	"github.com/jo-hoe/go-mail-service/app/validation"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Validator = &validation.GenericValidator{Validator: validator.New()}

	e.POST("/v1/sendmail", sendmailHandler)
	e.GET("/", probeHandler)

	secretService := secret.NewEnvSecretService()
	port, err := secretService.Get("API_PORT")
	if err != nil {
		e.Logger.Fatal(err)
	}

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}

func sendmailHandler(ctx echo.Context) (err error) {
	mailAttributes := new(mail.MailAttributes)
	if err = ctx.Bind(mailAttributes); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err = ctx.Validate(mailAttributes); err != nil {
		return err
	}

	secretService := secret.NewEnvSecretService()
	apiKey, err := secretService.Get("SENDGRID_API_KEY")
	if err != nil {
		return err
	}
	fromName, err := secretService.Get("DEFAULT_FROM_NAME")
	if err != nil {
		return err
	}
	fromAddress, err := secretService.Get("DEFAULT_FROM_ADDRESS")
	if err != nil {
		return err
	}

	mailService := sendgrid.NewSendGridService(&sendgrid.SendGridConfig{
		APIKey:        apiKey,
		OriginAddress: fromAddress,
		OriginName:    fromName,
	})
	if err = mailService.SendMail(ctx.Request().Context(), *mailAttributes); err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, mailAttributes)
}

func probeHandler(ctx echo.Context) (err error) {
	return ctx.NoContent(http.StatusOK)
}