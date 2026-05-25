package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator"
	"github.com/jo-hoe/go-mail-service/internal/mail"
	"github.com/jo-hoe/go-mail-service/internal/mail/noop"
	"github.com/jo-hoe/go-mail-service/internal/validation"
	"github.com/labstack/echo/v4"
)

// errorMailService always returns an error from SendMail.
type errorMailService struct{}

func (e *errorMailService) SendMail(_ context.Context, _ mail.MailAttributes) error {
	return echo.NewHTTPError(http.StatusInternalServerError, "send failed")
}

func Test_sendMailHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		svc        mail.MailService
		wantStatus int
		wantErr    bool
	}{
		{
			name:       "valid request",
			body:       `{"to": "test@example.com", "subject": "Test Subject", "content": "Test Body"}`,
			svc:        noop.NewNoopService(),
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "multiple recipients",
			body:       `{"to": "a@example.com,b@example.com", "subject": "Test", "content": "Body"}`,
			svc:        noop.NewNoopService(),
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:    "missing required field",
			body:    `{}`,
			svc:     noop.NewNoopService(),
			wantErr: true,
		},
		{
			name:    "invalid json",
			body:    `not json`,
			svc:     noop.NewNoopService(),
			wantErr: true,
		},
		{
			name:    "service error",
			body:    `{"to": "test@example.com", "subject": "Subject", "content": "Body"}`,
			svc:     &errorMailService{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newContextWithBody(tt.body)
			handler := sendMailHandler(tt.svc)
			err := handler(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("sendMailHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_probeHandler(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := probeHandler(ctx); err != nil {
		t.Errorf("probeHandler() error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("probeHandler() status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func newContextWithBody(body string) echo.Context {
	e := echo.New()
	e.Validator = &validation.GenericValidator{Validator: validator.New()}
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}
