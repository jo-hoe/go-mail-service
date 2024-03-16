package validator

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

type GenericValidator struct {
    validator *validator.Validate
}

func (gv *GenericValidator) Validate(i interface{}) error {
	if err := gv.validator.Struct(i); err != nil {
	  return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("received invalid request body: %v", err))
	}
	return nil
}
