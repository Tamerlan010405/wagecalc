package postgres

import (
	"context"
	"database/sql"
	"time"

	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/domain"
)

func insertResultSnapshot(ctx context.Context, tx *sql.Tx, calculation domain.Calculation, encryptionKey string) error {
	result := calculation.Result
	warningsJSON, err := warningsToJSON(result.Warnings)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		insert into result_snapshots (
			calculation_id,
			threshold_period_year,
			threshold_period_quarter,
			months_with_payments,
			monthly_average_sum_ciphertext,
			quarter_average_ciphertext,
			threshold_used_ciphertext,
			threshold_source,
			delta_ciphertext,
			delta_percent_ciphertext,
			verdict_ciphertext,
			recommendation_type_ciphertext,
			estimated_fund_change_ciphertext,
			warnings_ciphertext,
			engine_version,
			checked_at
		)
		values(
			$1, $2, $3, $4,
			pgp_sym_encrypt($5::text, $17, 'cipher-algo=aes256'),
			pgp_sym_encrypt($6::text, $17, 'cipher-algo=aes256'),
			case when $7::bigint is null then null else pgp_sym_encrypt($7::text, $17, 'cipher-algo=aes256') end,
			$8,
			case when $9::bigint is null then null else pgp_sym_encrypt($9::text, $17, 'cipher-algo=aes256') end,
			case when $10::text is null then null else pgp_sym_encrypt($10::text, $17, 'cipher-algo=aes256') end,
			pgp_sym_encrypt($11::text, $17, 'cipher-algo=aes256'),
			pgp_sym_encrypt($12::text, $17, 'cipher-algo=aes256'),
			pgp_sym_encrypt($13::text, $17, 'cipher-algo=aes256'),
			pgp_sym_encrypt($14::text, $17, 'cipher-algo=aes256'),
			$15, $16
		)
	`,
		calculation.ID,
		result.ThresholdPeriod.Year,
		int(result.ThresholdPeriod.Quarter),
		result.Quarter.MonthsWithPayments,
		result.Quarter.MonthlyAverageSum.GetCents(),
		result.Quarter.Average.GetCents(),
		moneyPtrToNullInt64(result.Threshold.Value),
		string(result.Threshold.Source),
		moneyPtrToNullInt64(result.Comparison.Delta),
		stringPtrToNullString(result.Comparison.DeltaPercent),
		string(result.Comparison.Verdict),
		string(result.Recommendation.Type),
		result.Recommendation.EstimatedQuarterFundChange.GetCents(),
		string(warningsJSON),
		calculation.EngineVersion,
		result.CheckedAt,
		encryptionKey,
	)
	return err
}

func getResultSnapshot(ctx context.Context, db *sql.DB, calculationID, encryptionKey string) (domain.CalculationResult, string, error) {
	row := db.QueryRowContext(ctx, `
		select
			threshold_period_year,
			threshold_period_quarter,
			months_with_payments,
			pgp_sym_decrypt(monthly_average_sum_ciphertext, $2)::bigint,
			pgp_sym_decrypt(quarter_average_ciphertext, $2)::bigint,
			pgp_sym_decrypt(threshold_used_ciphertext, $2)::bigint,
			threshold_source,
			pgp_sym_decrypt(delta_ciphertext, $2)::bigint,
			pgp_sym_decrypt(delta_percent_ciphertext, $2)::numeric,
			pgp_sym_decrypt(verdict_ciphertext, $2),
			pgp_sym_decrypt(recommendation_type_ciphertext, $2),
			pgp_sym_decrypt(estimated_fund_change_ciphertext, $2)::bigint,
			pgp_sym_decrypt(warnings_ciphertext, $2)::jsonb,
			engine_version,
			checked_at
		from result_snapshots
		where calculation_id = $1
	`, calculationID, encryptionKey)
	var thresholdPeriodYear int
	var thresholdPeriodQuarter int
	var monthsWithPayments int
	var monthlyAverageSumCents int64
	var quarterAverageCents int64
	var thresholdUsed sql.NullInt64
	var thresholdSource string
	var delta sql.NullInt64
	var deltaPercent sql.NullString
	var verdict string
	var recommendationType string
	var estimatedQuarterFundChange int64
	var warningsJSON []byte
	var engineVersion string
	var checkedAt time.Time
	if err := row.Scan(
		&thresholdPeriodYear,
		&thresholdPeriodQuarter,
		&monthsWithPayments,
		&monthlyAverageSumCents,
		&quarterAverageCents,
		&thresholdUsed,
		&thresholdSource,
		&delta,
		&deltaPercent,
		&verdict,
		&recommendationType,
		&estimatedQuarterFundChange,
		&warningsJSON,
		&engineVersion,
		&checkedAt,
	); err != nil {
		return domain.CalculationResult{}, "", errDecryptCalculationData
	}
	warnings, err := warningsFromJSON(warningsJSON)
	if err != nil {
		return domain.CalculationResult{}, "", errDecryptCalculationData
	}
	salaryChange := domain.NewMoney(0)
	deltaMoney := nullInt64ToMoneyPtr(delta)
	if deltaMoney != nil {
		salaryChange = deltaMoney.Abs()
	}
	result := domain.CalculationResult{
		ThresholdPeriod: domain.Period{Year: thresholdPeriodYear, Quarter: domain.Quarter(thresholdPeriodQuarter)},
		Quarter:         domain.QuarterResult{MonthsWithPayments: monthsWithPayments, MonthlyAverageSum: domain.NewMoney(monthlyAverageSumCents), Average: domain.NewMoney(quarterAverageCents)},
		Threshold:       domain.ResolvedThreshold{Value: nullInt64ToMoneyPtr(thresholdUsed), Source: domain.ThresholdSource(thresholdSource)},
		Comparison:      domain.ComparisonResult{Delta: deltaMoney, DeltaPercent: nullStringToStringPtr(deltaPercent), Verdict: domain.Verdict(verdict)},
		Recommendation:  domain.RecommendationResult{Type: domain.RecommendationType(recommendationType), SalaryChangePerEmployeePerMonth: salaryChange, EstimatedQuarterFundChange: domain.NewMoney(estimatedQuarterFundChange)},
		Warnings:        warnings,
		CheckedAt:       checkedAt,
	}
	return result, engineVersion, nil
}
