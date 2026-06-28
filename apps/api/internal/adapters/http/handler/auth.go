package handler

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
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

type loginResponse struct {
	UserID uuid.UUID `json:"user_id" swaggertype:"string" format:"uuid"`
	Name   string    `json:"name"`
}

// @Summary  Register a new user
// @Tags     auth
// @Accept   json
// @Produce  json
// @Param    payload  body      appauth.RegisterInput  true  "Registration payload"
// @Success  201      {object}  appauth.RegisterOutput
// @Failure  422      {object}  map[string]string  "validation error"
// @Router   /auth/register [post]
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

// @Summary  Log in
// @Tags     auth
// @Accept   json
// @Produce  json
// @Param    payload  body      appauth.LoginInput  true  "Login payload"
// @Success  200      {object}  loginResponse
// @Failure  401      {object}  map[string]string  "invalid credentials"
// @Failure  403      {object}  map[string]string  "not a member of the organization"
// @Router   /auth/login [post]
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

	return c.JSON(http.StatusOK, loginResponse{UserID: out.UserID, Name: out.Name})
}

// @Summary  Log out
// @Tags     auth
// @Success  204  "no content"
// @Router   /auth/logout [post]
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
