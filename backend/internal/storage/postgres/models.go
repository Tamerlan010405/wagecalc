package postgres

import (
	"database/sql"
	"encoding/json"

	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/domain"
)

func moneyPtrToNullInt64(money *domain.Money) sql.NullInt64 {
	if money == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{
		Int64: money.GetCents(),
		Valid: true,
	}
}

func nullInt64ToMoneyPtr(value sql.NullInt64) *domain.Money {
	if !value.Valid {
		return nil
	}
	money := domain.NewMoney(value.Int64)
	return &money
}

func stringPtrToNullString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{
		String: *value,
		Valid:  true,
	}
}

func nullStringToStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}

func warningsToJSON(warnings []domain.CalculationWarning) ([]byte, error) {
	if warnings == nil {
		warnings = []domain.CalculationWarning{}
	}
	return json.Marshal(warnings)
}

func warningsFromJSON(data []byte) ([]domain.CalculationWarning, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var warnings []domain.CalculationWarning
	if err := json.Unmarshal(data, &warnings); err != nil {
		return nil, err
	}

	return warnings, nil
}
