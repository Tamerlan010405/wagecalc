package postgres

import (
	"context"
	"database/sql"

	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/domain"
)

func insertMonthlyResults(ctx context.Context, tx *sql.Tx, calculationID string, monthlyResults []domain.MonthlyResult, encryptionKey string) error {
	for _, result := range monthlyResults {
		_, err := tx.ExecContext(ctx, `
			insert into monthly_results (
				calculation_id,
				month_index,
				labor_payment_sum_ciphertext,
				headcount_ciphertext,
				average_ciphertext,
				included_in_quarter_average
			)
			values (
				$1, $2,
				pgp_sym_encrypt($3::text, $7, 'cipher-algo=aes256'),
				pgp_sym_encrypt($4::text, $7, 'cipher-algo=aes256'),
				pgp_sym_encrypt($5::text, $7, 'cipher-algo=aes256'),
				$6
			)
		`,
			calculationID,
			result.MonthIndex,
			result.LaborPaymentSum.GetCents(),
			result.Headcount,
			result.Average.GetCents(),
			result.IncludedInQuarterAverage,
			encryptionKey,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func getMonthlyResults(ctx context.Context, db *sql.DB, calculationID, encryptionKey string) ([]domain.MonthlyResult, error) {
	rows, err := db.QueryContext(ctx, `
		select
			month_index,
			pgp_sym_decrypt(labor_payment_sum_ciphertext, $2)::bigint,
			pgp_sym_decrypt(headcount_ciphertext, $2)::integer,
			pgp_sym_decrypt(average_ciphertext, $2)::bigint,
			included_in_quarter_average
		from monthly_results
		where calculation_id = $1
		order by month_index
	`, calculationID, encryptionKey)
	if err != nil {
		return nil, errDecryptCalculationData
	}
	defer rows.Close()
	monthlyResults := make([]domain.MonthlyResult, 0, domain.CountMonth)
	for rows.Next() {
		var monthIndex int
		var laborPaymentSum int64
		var headcount int
		var average int64
		var includedInQuarterAverage bool
		if err := rows.Scan(
			&monthIndex,
			&laborPaymentSum,
			&headcount,
			&average,
			&includedInQuarterAverage,
		); err != nil {
			return nil, errDecryptCalculationData
		}
		monthlyResult := domain.MonthlyResult{
			MonthIndex:               monthIndex,
			LaborPaymentSum:          domain.NewMoney(laborPaymentSum),
			Headcount:                headcount,
			Average:                  domain.NewMoney(average),
			IncludedInQuarterAverage: includedInQuarterAverage,
		}
		monthlyResults = append(monthlyResults, monthlyResult)
	}
	if err := rows.Err(); err != nil {
		return nil, errDecryptCalculationData
	}
	return monthlyResults, nil
}
