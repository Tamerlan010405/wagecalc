package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/domain"
	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/repository"
)

type calculationBase struct {
	ID              string
	OrgName         string
	Title           string
	PeriodYear      int
	PeriodQuarter   int
	CreatedByUserID string
	CreatedAt       time.Time
}

func insertCalculation(ctx context.Context, tx *sql.Tx, calculation domain.Calculation, organizationID string) error {
	_, err := tx.ExecContext(ctx, `
		insert into calculations (
			id,
			organization_id,
			title,
			period_year,
			period_quarter,
			created_by_user_id, 
			created_at
		)
		values ($1, $2, $3, $4, $5, $6, $7)
	`,
		calculation.ID,
		organizationID,
		calculation.Input.Title,
		calculation.Input.Period.Year,
		int(calculation.Input.Period.Quarter),
		calculation.CreatedByUserID,
		calculation.CreatedAt,
	)

	return err
}

func getCalculationBase(ctx context.Context, db *sql.DB, userID, id string) (calculationBase, error) {
	var base calculationBase
	err := db.QueryRowContext(ctx, `
		select
			c.id,
			o.name,
			c.title,
			c.period_year,
			c.period_quarter,
			c.created_by_user_id,
			c.created_at
		from calculations c
		join organizations o on o.id = c.organization_id
		join organization_members om on om.organization_id = o.id
		where c.id = $1
		and om.user_id = $2
	`, id, userID).Scan(
		&base.ID,
		&base.OrgName,
		&base.Title,
		&base.PeriodYear,
		&base.PeriodQuarter,
		&base.CreatedByUserID,
		&base.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return calculationBase{}, repository.ErrCalculationNotFound
	}
	if err != nil {
		return calculationBase{}, err
	}

	return base, nil
}
