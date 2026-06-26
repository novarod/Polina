package handler

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	appauth "github.com/novarod/polina/apps/api/internal/application/auth"
)

type AuthHandler struct {
	register *appauth.RegisterUseCase
	login    *appauth.LoginUseCase
}

func NewAuthHandler(register *appauth.RegisterUseCase, login *appauth.LoginUseCase) *AuthHandler {
	return &AuthHandler{register: register, login: login}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var in appauth.RegisterInput
	if err := bindAndValidate(c, &in); err != nil {
		return err
	}
	out, err := h.register.Execute(c.Request().Context(), in)
	if err != nil {
		return mapError(err)
	}
	return c.JSON(http.StatusCreated, out)
}

func (h *AuthHandler) Login(c echo.Context) error {
	var in appauth.LoginInput
	if err := bindAndValidate(c, &in); err != nil {
		return err
	}
	out, err := h.login.Execute(c.Request().Context(), in)
	if err != nil {
		return mapError(err)
	}

	secure := os.Getenv("ENV") == "production"
	expiryH, _ := strconv.Atoi(os.Getenv("JWT_EXPIRY_HOURS"))
	if expiryH == 0 {
		expiryH = 24
	}

	c.SetCookie(&http.Cookie{
		Name:     "session",
		Value:    out.Token,
		Path:     "/",
		Expires:  time.Now().Add(time.Duration(expiryH) * time.Hour),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	})

	return c.JSON(http.StatusOK, map[string]any{
		"user_id": out.UserID,
		"name":    out.Name,
	})
}

func (h *AuthHandler) Logout(c echo.Context) error {
	secure := os.Getenv("ENV") == "production"
	c.SetCookie(&http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	})
	return c.NoContent(http.StatusNoContent)
}
