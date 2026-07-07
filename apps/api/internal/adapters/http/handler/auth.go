package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	httpmw "github.com/novarod/polina/apps/api/internal/adapters/http/middleware"
	appauth "github.com/novarod/polina/apps/api/internal/application/auth"
)

type CookieConfig struct {
	Secure      bool
	ExpiryHours int
}

type AuthHandler struct {
	register  *appauth.RegisterUseCase
	login     *appauth.LoginUseCase
	logoutAll *appauth.LogoutAllUseCase
	me        *appauth.MeUseCase
	cookie    CookieConfig
}

func NewAuthHandler(register *appauth.RegisterUseCase, login *appauth.LoginUseCase, logoutAll *appauth.LogoutAllUseCase, me *appauth.MeUseCase, cookie CookieConfig) *AuthHandler {
	return &AuthHandler{register: register, login: login, logoutAll: logoutAll, me: me, cookie: cookie}
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

	c.SetCookie(&http.Cookie{
		Name:     "session",
		Value:    out.Token,
		Path:     "/",
		Expires:  time.Now().Add(time.Duration(h.cookie.ExpiryHours) * time.Hour),
		HttpOnly: true,
		Secure:   h.cookie.Secure,
		SameSite: http.SameSiteStrictMode,
	})

	return c.JSON(http.StatusOK, loginResponse{UserID: out.UserID, Name: out.Name})
}

// @Summary  Current session user
// @Tags     auth
// @Produce  json
// @Security BearerAuth
// @Success  200  {object}  appauth.MeOutput
// @Failure  401  {object}  map[string]string  "missing or invalid session"
// @Router   /auth/me [get]
func (h *AuthHandler) Me(c echo.Context) error {
	claims := httpmw.MustGetClaims(c)
	out, err := h.me.Execute(c.Request().Context(), claims.UserID)
	if err != nil {
		return mapError(err)
	}
	return c.JSON(http.StatusOK, out)
}

// @Summary  Log out
// @Tags     auth
// @Success  204  "no content"
// @Router   /auth/logout [post]
func (h *AuthHandler) Logout(c echo.Context) error {
	h.clearSessionCookie(c)
	return c.NoContent(http.StatusNoContent)
}

// @Summary  Log out everywhere (revoke all sessions)
// @Tags     auth
// @Security BearerAuth
// @Success  204  "no content"
// @Failure  401  {object}  map[string]string  "missing or invalid session"
// @Router   /auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c echo.Context) error {
	claims := httpmw.MustGetClaims(c)
	if err := h.logoutAll.Execute(c.Request().Context(), claims.UserID); err != nil {
		return mapError(err)
	}
	h.clearSessionCookie(c)
	return c.NoContent(http.StatusNoContent)
}

func (h *AuthHandler) clearSessionCookie(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cookie.Secure,
		SameSite: http.SameSiteStrictMode,
	})
}
