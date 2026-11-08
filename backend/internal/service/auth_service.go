package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/domain"
	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/repository"
)

const (
	sessionTTL                   = 7 * 24 * time.Hour
	passwordResetTTL             = 15 * time.Minute
	passwordResetMaxAttempts     = 5
	passwordResetRequestCooldown = 5 * time.Minute
)

type EmailSender interface {
	SendPasswordResetCode(ctx context.Context, to string, code string) error
}

type AuthService struct {
	repo                repository.AuthRepository
	emailSender         EmailSender
	passwordResetSecret string
	now                 func() time.Time
}

type AuthResult struct {
	User         domain.User
	SessionToken string
	Session      domain.Session
}

func NewAuthService(repo repository.AuthRepository, emailSender EmailSender, passwordResetSecret string) *AuthService {
	return &AuthService{
		repo:                repo,
		emailSender:         emailSender,
		passwordResetSecret: passwordResetSecret,
		now:                 time.Now,
	}
}

func (s *AuthService) Register(ctx context.Context, email string, password string, displayName *string) (AuthResult, error) {
	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return AuthResult{}, err
	}
	if err := validatePassword(password); err != nil {
		return AuthResult{}, err
	}
	normalizedDisplayName := normalizeOptionalString(displayName)
	now := s.now()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResult{}, err
	}
	user := domain.User{
		Email:       normalizedEmail,
		DisplayName: normalizedDisplayName,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	createdUser, err := s.repo.CreatePasswordUser(ctx, user, string(passwordHash))
	if err != nil {
		return AuthResult{}, err
	}
	return s.createSession(ctx, createdUser, now)
}

func (s *AuthService) createSession(ctx context.Context, user domain.User, now time.Time) (AuthResult, error) {
	token, err := newSessionToken()
	if err != nil {
		return AuthResult{}, err
	}
	session := domain.Session{
		UserID:    user.ID,
		TokenHash: hashSessionToken(token),
		ExpiresAt: now.Add(sessionTTL),
		CreatedAt: now,
	}
	createdSession, err := s.repo.CreateSession(ctx, session)
	if err != nil {
		return AuthResult{}, err
	}
	return AuthResult{
		User:         user,
		SessionToken: token,
		Session:      createdSession,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, email string, password string) (AuthResult, error) {
	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return AuthResult{}, err
	}
	user, identity, err := s.repo.GetPasswordIdentityByEmail(ctx, normalizedEmail)
	if err != nil {
		return AuthResult{}, err
	}
	if identity.PasswordHash == nil {
		return AuthResult{}, repository.ErrUserNotFound
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*identity.PasswordHash), []byte(password)); err != nil {
		return AuthResult{}, repository.ErrUserNotFound
	}
	return s.createSession(ctx, user, s.now())
}

func (s *AuthService) Logout(ctx context.Context, sessionToken string) error {
	tokenHash := hashSessionToken(sessionToken)
	return s.repo.DeleteSessionByTokenHash(ctx, tokenHash)
}

func (s *AuthService) Me(ctx context.Context, sessionToken string) (domain.User, error) {
	tokenHash := hashSessionToken(sessionToken)
	user, _, err := s.repo.GetUserBySessionTokenHash(ctx, tokenHash)
	return user, err
}

func normalizeEmail(email string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return "", domain.NewDomainError(domain.ErrorInvalidEmail, "email", "email is required")
	}
	if _, err := mail.ParseAddress(normalized); err != nil {
		return "", domain.NewDomainError(domain.ErrorInvalidEmail, "email", "email is invalid")
	}
	return normalized, nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return domain.NewDomainError(domain.ErrorInvalidPassword, "password", "password must contain at least 8 characters")
	}
	if len(password) > 72 {
		return domain.NewDomainError(domain.ErrorInvalidPassword, "password", "password must contain no more than 72 bytes")
	}
	return nil
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil
	}
	return &normalized
}

func newSessionToken() (string, error) {
	var bytes [32]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes[:]), nil
}

func hashSessionToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func IsAuthNotFound(err error) bool {
	return errors.Is(err, repository.ErrUserNotFound) || errors.Is(err, repository.ErrSessionNotFound)
}

func (s *AuthService) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	return s.repo.DeleteExpiredSessions(ctx, s.now())
}
