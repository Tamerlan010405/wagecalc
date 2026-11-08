package repository

import (
	"context"
	"errors"
	"time"

	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/domain"
)

var (
	ErrUserNotFound              = errors.New("user not found")
	ErrAuthIdentityFound         = errors.New("auth identity already exists")
	ErrSessionNotFound           = errors.New("session not found")
	ErrPasswordResetCodeNotFound = errors.New("password reset code not found")
)

type AuthRepository interface {
	CreatePasswordUser(ctx context.Context, user domain.User, passwordHash string) (domain.User, error)
	GetPasswordIdentityByEmail(ctx context.Context, email string) (domain.User, domain.AuthIdentity, error)
	CreateSession(ctx context.Context, session domain.Session) (domain.Session, error)
	GetUserBySessionTokenHash(ctx context.Context, tokenHash string) (domain.User, domain.Session, error)
	DeleteSessionByTokenHash(ctx context.Context, tokenHash string) error
	DeleteExpiredSessions(ctx context.Context, now time.Time) (int64, error)
	CreatePasswordResetCode(ctx context.Context, resetCode domain.PasswordResetCode) error
	GetActivePasswordResetCodeByUserID(ctx context.Context, userID string) (domain.PasswordResetCode, error)
	IncrementPasswordResetCodeAttempts(ctx context.Context, resetCodeID string) error
	CompletePasswordReset(ctx context.Context, userID string, resetCodeID string, passwordHash string, now time.Time) error
}
