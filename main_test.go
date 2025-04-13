package main

import (
    "bytes"
    "net/http"
    "net/http/httptest"
    "os"
    "testing"

    "github.com/go-playground/validator"
    "github.com/jo-hoe/go-mail-service/app/validation"
    "github.com/labstack/echo/v4"
)

func Test_sendMailHandler(t *testing.T) {
    if err := os.Setenv(IS_NOOP_ENABLED_ENV_KEY, "true"); err != nil {
        t.Fatalf("Failed to set environment variable: %v", err)
    }
    defer func() {
        if err := os.Unsetenv(IS_NOOP_ENABLED_ENV_KEY); err != nil {
            t.Fatalf("Failed to unset environment variable: %v", err)
        }
    }()
    if err := os.Setenv(IS_SENDGRID_ENABLED_ENV_KEY, "false"); err != nil {
        t.Fatalf("Failed to set environment variable: %v", err)
    }
    defer func() {
        if err := os.Unsetenv(IS_SENDGRID_ENABLED_ENV_KEY); err != nil {
            t.Fatalf("Failed to unset environment variable: %v", err)
        }
    }()

    type args struct {
        ctx echo.Context
    }
    tests := []struct {
        name    string
        args    args
        wantErr bool
    }{
        {
            name: "handle basic mail",
            args: args{
                ctx: newContextWithBody(`{"to": "test@example.com,test2@example.com", "subject": "Test Subject", "content": "Test Body"}`),
            },
            wantErr: false,
        },
        {
            name: "missing field",
            args: args{
                ctx: newContextWithBody(`{}`),
            },
            wantErr: true,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if err := sendMailHandler(tt.args.ctx); (err != nil) != tt.wantErr {
                t.Errorf("sendMailHandler() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}

func newContextWithBody(body string) echo.Context {
    e := echo.New()
    e.Validator = &validation.GenericValidator{Validator: validator.New()}
    req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))
    req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)
    return c
}