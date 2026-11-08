package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/domain"
	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/repository"
)

func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return err
	}

	user, _, err := s.repo.GetPasswordIdentityByEmail(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil
		}
		return err
	}
	now := s.now()

	activeCode, err := s.repo.GetActivePasswordResetCodeByUserID(ctx, user.ID)
	if err != nil && !errors.Is(err, repository.ErrPasswordResetCodeNotFound) {
		return err
	}
	if err == nil && activeCode.CreatedAt.Add(passwordResetRequestCooldown).After(now) {
		return nil
	}

	code, err := newPasswordResetCode()
	if err != nil {
		return err
	}

	resetCode := domain.PasswordResetCode{
		UserID:    user.ID,
		CodeHash:  s.hashPasswordResetCode(normalizedEmail, code),
		ExpiresAt: now.Add(passwordResetTTL),
		Attempts:  0,
		CreatedAt: now,
	}

	if err := s.repo.CreatePasswordResetCode(ctx, resetCode); err != nil {
		return err
	}

	if s.emailSender == nil {
		return nil
	}

	return s.emailSender.SendPasswordResetCode(ctx, normalizedEmail, code)
}

func (s *AuthService) ConfirmPasswordReset(ctx context.Context, email string, code string, newPassword string) error {
	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return err
	}

	normalizedCode, err := normalizePasswordResetCode(code)
	if err != nil {
		return err
	}

	if err := validatePassword(newPassword); err != nil {
		return err
	}

	user, _, err := s.repo.GetPasswordIdentityByEmail(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return repository.ErrPasswordResetCodeNotFound
		}
		return err
	}

	resetCode, err := s.repo.GetActivePasswordResetCodeByUserID(ctx, user.ID)
	if err != nil {
		return err
	}

	now := s.now()
	if !resetCode.ExpiresAt.After(now) {
		return repository.ErrPasswordResetCodeNotFound
	}

	if resetCode.Attempts >= passwordResetMaxAttempts {
		return repository.ErrPasswordResetCodeNotFound
	}

	expectedHash := s.hashPasswordResetCode(normalizedEmail, normalizedCode)
	if !hmac.Equal([]byte(resetCode.CodeHash), []byte(expectedHash)) {
		if err := s.repo.IncrementPasswordResetCodeAttempts(ctx, resetCode.ID); err != nil &&
			!errors.Is(err, repository.ErrPasswordResetCodeNotFound) {
			return err
		}
		return repository.ErrPasswordResetCodeNotFound
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repo.CompletePasswordReset(ctx, user.ID, resetCode.ID, string(passwordHash), now)
}

func newPasswordResetCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func normalizePasswordResetCode(code string) (string, error) {
	normalized := strings.TrimSpace(code)
	if len(normalized) != 6 {
		return "", domain.NewDomainError(domain.ErrorInvalidPasswordResetCode, "code", "code must contain 6 digits")
	}

	for _, char := range normalized {
		if char < '0' || char > '9' {
			return "", domain.NewDomainError(domain.ErrorInvalidPasswordResetCode, "code", "code must contain only digits")
		}
	}

	return normalized, nil
}

func (s *AuthService) hashPasswordResetCode(email string, code string) string {
	mac := hmac.New(sha256.New, []byte(s.passwordResetSecret))
	mac.Write([]byte(strings.ToLower(strings.TrimSpace(email))))
	mac.Write([]byte(":"))
	mac.Write([]byte(code))
	return hex.EncodeToString(mac.Sum(nil))
}
