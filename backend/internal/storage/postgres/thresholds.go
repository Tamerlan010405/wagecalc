package postgres

import (
	"context"
	"database/sql"

	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/domain"
)

const thresholdSourceManual = "MANUAL"

func insertThreshold(ctx context.Context, tx *sql.Tx, calculationID string, threshold domain.Threshold) error {
	_, err := tx.ExecContext(ctx, `
		insert into thresholds (
			calculation_id,
			value_rf_cents,
			value_region_cents,
			region_code,
			source
		)
		values ($1, $2, $3, $4, $5)
	`,
		calculationID,
		moneyPtrToNullInt64(threshold.ValueRF),
		moneyPtrToNullInt64(threshold.ValueRegion),
		stringPtrToNullString(threshold.RegionCode),
		thresholdSourceManual,
	)
	return err
}

func getThreshold(ctx context.Context, db *sql.DB, calculationID string) (domain.Threshold, error) {
	var valueRF sql.NullInt64
	var valueRegion sql.NullInt64
	var regionCode sql.NullString
	err := db.QueryRowContext(ctx, `
		select
			value_rf_cents,
			value_region_cents,
			region_code
		from thresholds
		where calculation_id = $1
	`, calculationID).Scan(
		&valueRF,
		&valueRegion,
		&regionCode,
	)
	if err != nil {
		return domain.Threshold{}, err
	}
	return domain.Threshold{
		ValueRF:     nullInt64ToMoneyPtr(valueRF),
		ValueRegion: nullInt64ToMoneyPtr(valueRegion),
		RegionCode:  nullStringToStringPtr(regionCode),
	}, nil
}
