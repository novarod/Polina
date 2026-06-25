package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

func bindAndValidate(c echo.Context, v any) error {
	if err := c.Bind(v); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(v); err != nil {
		return err
	}
	return nil
}

func mapError(err error) error {
	var appErr *apierr.AppError
	if errors.As(err, &appErr) {
		return echo.NewHTTPError(appErr.Code, appErr.Message)
	}
	return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
}
