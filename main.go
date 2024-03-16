package main

import (
	"fmt"
	"net/http"

	"github.com/jo-hoe/go-sendgrid-service/app/secret"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", hello)

	secretService := secret.NewEnvSecretService()
	port, err := secretService.Get("API_PORT")
	if err != nil {
		e.Logger.Fatal(err)
	}

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}

func hello(context echo.Context) error {
	return context.String(http.StatusOK, "Hello, World!")
}
