package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

func getOrCreateOrganization(ctx context.Context, tx *sql.Tx, userID, name string, createdAt time.Time) (string, error) {
	normalizedName := normalizeOrganizationName(name)
	var id string
	err := tx.QueryRowContext(ctx, `
		select id
		from organizations o
		join organization_members om on om.organization_id = o.id
		where om.user_id = $1
		and om.organization_name_key = lower($2)
	`, userID, normalizedName).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	err = tx.QueryRowContext(ctx, `
		insert into organizations (name, normalized_name)
		values ($1, $2)
		returning id
	`, name, normalizedName).Scan(&id)

	if err != nil {
		return "", err
	}

	_, err = tx.ExecContext(ctx, `
		insert into organization_members (
			organization_id,
			user_id,
			role,
			organization_name_key,
			created_at
		)
		values ($1, $2, 'OWNER', $3, $4)
	`, id, userID, normalizedName, createdAt)
	if err != nil {
		return "", err
	}

	return id, nil
}

func normalizeOrganizationName(name string) string {
	return strings.ToLower(strings.Join(strings.Fields(name), " "))
}
