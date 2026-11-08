package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/domain"
	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/repository"
)

const passwordProvider = "password"

type AuthRepository struct {
	db *sql.DB
}

var _ repository.AuthRepository = (*AuthRepository)(nil)

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) CreatePasswordUser(ctx context.Context, user domain.User, passwordHash string) (domain.User, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.User{}, err
	}
	defer tx.Rollback()
	var created domain.User
	var displayName sql.NullString
	err = tx.QueryRowContext(ctx, `
		insert into users (
			email,
			display_name,
			created_at,
			updated_at
		)
		values ($1, $2, $3, $4)
		returning id, email, display_name, created_at, updated_at
	`,
		user.Email,
		stringPtrToNullString(user.DisplayName),
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(
		&created.ID,
		&created.Email,
		&displayName,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.User{}, repository.ErrAuthIdentityFound
		}
		return domain.User{}, err
	}

	created.DisplayName = nullStringToStringPtr(displayName)

	_, err = tx.ExecContext(ctx, `
		insert into auth_identities (
			user_id,
			provider,
			provider_user_id,
			password_hash,
			created_at
		)
		values ($1, $2, $3, $4, $5)
	`,
		created.ID,
		passwordProvider,
		created.Email,
		passwordHash,
		user.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.User{}, repository.ErrAuthIdentityFound
		}
		return domain.User{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.User{}, err
	}
	return created, nil
}

func (r *AuthRepository) GetPasswordIdentityByEmail(ctx context.Context, email string) (domain.User, domain.AuthIdentity, error) {
	var user domain.User
	var identity domain.AuthIdentity
	var displayName sql.NullString
	var passwordHash sql.NullString

	err := r.db.QueryRowContext(ctx, `
		select
			u.id,
			u.email,
			u.display_name,
			u.created_at,
			u.updated_at,
			ai.id,
			ai.user_id,
			ai.provider,
			ai.provider_user_id,
			ai.password_hash,
			ai.created_at
		from auth_identities ai
		join users u on u.id = ai.user_id
		where ai.provider = $1
		and lower(ai.provider_user_id) = lower($2)
	`, passwordProvider, email).Scan(
		&user.ID,
		&user.Email,
		&displayName,
		&user.CreatedAt,
		&user.UpdatedAt,
		&identity.ID,
		&identity.UserID,
		&identity.Provider,
		&identity.ProviderUserID,
		&passwordHash,
		&identity.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, domain.AuthIdentity{}, repository.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, domain.AuthIdentity{}, err
	}
	user.DisplayName = nullStringToStringPtr(displayName)
	identity.PasswordHash = nullStringToStringPtr(passwordHash)
	return user, identity, nil
}

func (r *AuthRepository) CreateSession(ctx context.Context, session domain.Session) (domain.Session, error) {
	var created domain.Session
	err := r.db.QueryRowContext(ctx, `
		insert into sessions (
			user_id,
			token_hash,
			expires_at,
			created_at
		)
		values ($1, $2, $3, $4)
		returning id, user_id, token_hash, expires_at, created_at
	`,
		session.UserID,
		session.TokenHash,
		session.ExpiresAt,
		session.CreatedAt,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.TokenHash,
		&created.ExpiresAt,
		&created.CreatedAt,
	)
	if err != nil {
		return domain.Session{}, err
	}
	return created, nil
}

func (r *AuthRepository) GetUserBySessionTokenHash(ctx context.Context, tokenHash string) (domain.User, domain.Session, error) {
	var user domain.User
	var session domain.Session
	var displayName sql.NullString
	err := r.db.QueryRowContext(ctx, `
		select
			u.id,
			u.email,
			u.display_name,
			u.created_at,
			u.updated_at,
			s.id,
			s.user_id,
			s.token_hash,
			s.expires_at,
			s.created_at
		from sessions s
		join users u on u.id = s.user_id
		where s.token_hash = $1
		and s.expires_at > now()
	`, tokenHash).Scan(
		&user.ID,
		&user.Email,
		&displayName,
		&user.CreatedAt,
		&user.UpdatedAt,
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.ExpiresAt,
		&session.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, domain.Session{}, repository.ErrSessionNotFound
	}
	if err != nil {
		return domain.User{}, domain.Session{}, err
	}
	user.DisplayName = nullStringToStringPtr(displayName)
	return user, session, nil
}

func (r *AuthRepository) DeleteSessionByTokenHash(ctx context.Context, tokenHash string) error {
	result, err := r.db.ExecContext(ctx, `
		delete from sessions
		where token_hash = $1
	`, tokenHash)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return repository.ErrSessionNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

func (r *AuthRepository) DeleteExpiredSessions(ctx context.Context, now time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		delete from sessions
		where expires_at <= $1
	`, now)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *AuthRepository) CreatePasswordResetCode(ctx context.Context, resetCode domain.PasswordResetCode) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		update password_reset_codes
		set used_at = $2
		where user_id = $1
		and used_at is null
	`, resetCode.UserID, resetCode.CreatedAt)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		insert into password_reset_codes (
			user_id,
			code_hash,
			expires_at,
			attempts,
			created_at
		)
		values ($1, $2, $3, $4, $5)
	`,
		resetCode.UserID,
		resetCode.CodeHash,
		resetCode.ExpiresAt,
		resetCode.Attempts,
		resetCode.CreatedAt,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *AuthRepository) GetActivePasswordResetCodeByUserID(ctx context.Context, userID string) (domain.PasswordResetCode, error) {
	var resetCode domain.PasswordResetCode
	err := r.db.QueryRowContext(ctx, `
		select
			id,
			user_id,
			code_hash,
			expires_at,
			attempts,
			created_at
		from password_reset_codes
		where user_id = $1
		and used_at is null
		order by created_at desc
		limit 1
	`, userID).Scan(
		&resetCode.ID,
		&resetCode.UserID,
		&resetCode.CodeHash,
		&resetCode.ExpiresAt,
		&resetCode.Attempts,
		&resetCode.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PasswordResetCode{}, repository.ErrPasswordResetCodeNotFound
	}
	if err != nil {
		return domain.PasswordResetCode{}, err
	}

	return resetCode, nil
}

func (r *AuthRepository) IncrementPasswordResetCodeAttempts(ctx context.Context, resetCodeID string) error {
	result, err := r.db.ExecContext(ctx, `
		update password_reset_codes
		set attempts = attempts + 1
		where id = $1
		and used_at is null
		and expires_at > now()
		and attempts < 5
	`, resetCodeID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return repository.ErrPasswordResetCodeNotFound
	}
	return nil
}

func (r *AuthRepository) CompletePasswordReset(ctx context.Context, userID string, resetCodeID string, passwordHash string, now time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		update password_reset_codes
		set used_at = $3
		where id = $1
		and user_id = $2
		and used_at is null
		and expires_at > $3
		and attempts < 5
	`, resetCodeID, userID, now)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return repository.ErrPasswordResetCodeNotFound
	}

	_, err = tx.ExecContext(ctx, `
		update auth_identities
		set password_hash = $2
		where user_id = $1
		and provider = $3
	`, userID, passwordHash, passwordProvider)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		update users
		set updated_at = $2
		where id = $1
	`, userID, now)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		delete from sessions
		where user_id = $1
	`, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
