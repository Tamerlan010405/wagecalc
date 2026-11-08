package httpapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/domain"
	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/repository"
	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/service"
)

const sessionCookieName = "wagecalc_session"

type AuthService interface {
	Register(ctx context.Context, email string, password string, displayName *string) (service.AuthResult, error)
	Login(ctx context.Context, email string, password string) (service.AuthResult, error)
	Logout(ctx context.Context, sessionToken string) error
	Me(ctx context.Context, sessionToken string) (domain.User, error)
	RequestPasswordReset(ctx context.Context, email string) error
	ConfirmPasswordReset(ctx context.Context, email string, code string, newPassword string) error
}

type AuthHandler struct {
	service      AuthService
	cookieSecure bool
}

func NewAuthHandler(service AuthService, cookieSecure bool) *AuthHandler {
	return &AuthHandler{service: service, cookieSecure: cookieSecure}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	if h == nil || h.service == nil {
		writeInternalError(w)
		return
	}
	var requestDTO AuthRegisterRequestDTO
	if err := decodeJSON(r, &requestDTO); err != nil {
		writeError(w, err)
		return
	}
	result, err := h.service.Register(r.Context(), requestDTO.Email, requestDTO.Password, requestDTO.DisplayName)
	if err != nil {
		writeError(w, mapAuthError(err))
		return
	}
	setSessionCookie(w, result.SessionToken, result.Session.ExpiresAt, h.cookieSecure)
	writeJSON(w, http.StatusCreated, authResponseToDTO(result.User))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	if h == nil || h.service == nil {
		writeInternalError(w)
		return
	}
	var requestDTO AuthLoginRequestDTO
	if err := decodeJSON(r, &requestDTO); err != nil {
		writeError(w, err)
		return
	}
	result, err := h.service.Login(r.Context(), requestDTO.Email, requestDTO.Password)
	if err != nil {
		writeError(w, mapAuthError(err))
		return
	}
	setSessionCookie(w, result.SessionToken, result.Session.ExpiresAt, h.cookieSecure)
	writeJSON(w, http.StatusOK, authResponseToDTO(result.User))
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	if h == nil || h.service == nil {
		writeInternalError(w)
		return
	}
	token, ok := sessionTokenFromRequest(r)
	if ok {
		if err := h.service.Logout(r.Context(), token); err != nil && !errors.Is(err, repository.ErrSessionNotFound) {
			writeError(w, err)
			return
		}
	}
	clearSessionCookie(w, h.cookieSecure)
	writeNoContent(w)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	if h == nil || h.service == nil {
		writeInternalError(w)
		return
	}
	token, ok := sessionTokenFromRequest(r)
	if !ok {
		writeError(w, unauthorizedError())
		return
	}
	user, err := h.service.Me(r.Context(), token)
	if err != nil {
		writeError(w, mapAuthError(err))
		return
	}
	writeJSON(w, http.StatusOK, authResponseToDTO(user))
}

func sessionTokenFromRequest(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return "", false
	}
	return cookie.Value, true
}

func setSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time, cookieSecure bool) {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   cookieSecure,
	})
}

func clearSessionCookie(w http.ResponseWriter, cookieSecure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   cookieSecure,
	})
}

func unauthorizedError() *apiError {
	return newAPIError(http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", "", nil)
}

func conflictError() *apiError {
	return newAPIError(http.StatusConflict, "AUTH_IDENTITY_EXISTS", "user already exists", "email", nil)
}

func mapAuthError(err error) error {
	if errors.Is(err, repository.ErrAuthIdentityFound) {
		return conflictError()
	}
	if errors.Is(err, repository.ErrPasswordResetCodeNotFound) {
		return invalidPasswordResetCodeError()
	}
	var domainErr domain.DomainError
	if errors.As(err, &domainErr) && domainErr.Code == domain.ErrorInvalidPasswordResetCode {
		return invalidPasswordResetCodeError()
	}
	if service.IsAuthNotFound(err) {
		return unauthorizedError()
	}
	return err
}

func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	if h == nil || h.service == nil {
		writeInternalError(w)
		return
	}

	var requestDTO PasswordResetRequestDTO
	if err := decodeJSON(r, &requestDTO); err != nil {
		writeError(w, err)
		return
	}

	if err := h.service.RequestPasswordReset(
		r.Context(),
		requestDTO.Email,
	); err != nil {
		writeError(w, mapAuthError(err))
		return
	}

	writeNoContent(w)
}

func (h *AuthHandler) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	if h == nil || h.service == nil {
		writeInternalError(w)
		return
	}

	var requestDTO PasswordResetConfirmDTO
	if err := decodeJSON(r, &requestDTO); err != nil {
		writeError(w, err)
		return
	}

	if err := h.service.ConfirmPasswordReset(
		r.Context(),
		requestDTO.Email,
		requestDTO.Code,
		requestDTO.NewPassword,
	); err != nil {
		writeError(w, mapAuthError(err))
		return
	}

	writeNoContent(w)
}

func invalidPasswordResetCodeError() *apiError {
	return newAPIError(
		http.StatusUnprocessableEntity,
		"INVALID_PASSWORD_RESET_CODE",
		"invalid or expired password reset code",
		"code",
		nil,
	)
}
