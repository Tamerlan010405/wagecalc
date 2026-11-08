package postgres

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strconv"
	"time"

	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/domain"
	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/repository"
)

type CalculationRepository struct {
	db            *sql.DB
	encryptionKey string
}

var errDecryptCalculationData = errors.New("decrypt calculation data failed")
var errInvalidEncryptedData = errors.New("encrypted calculation data is invalid")

var _ repository.CalculationRepository = (*CalculationRepository)(nil)

func NewCalculationRepository(db *sql.DB, encryptionKey string) *CalculationRepository {
	return &CalculationRepository{
		db:            db,
		encryptionKey: encryptionKey,
	}
}

func (r *CalculationRepository) Create(ctx context.Context, calculation domain.Calculation) error {
	if err := validateEncryptedCalculation(calculation); err != nil {
		return err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	organizationID, err := getOrCreateOrganization(ctx, tx, calculation.CreatedByUserID, calculation.Input.OrgName, calculation.CreatedAt)
	if err != nil {
		return err
	}
	if err := insertCalculation(ctx, tx, calculation, organizationID); err != nil {
		return err
	}
	if err := insertThreshold(ctx, tx, calculation.ID, calculation.Input.Threshold); err != nil {
		return err
	}
	if err := insertPayrollEntries(ctx, tx, calculation.ID, calculation.Input.Employees, r.encryptionKey); err != nil {
		return err
	}
	if err := insertMonthlyResults(ctx, tx, calculation.ID, calculation.Result.MonthlyResults, r.encryptionKey); err != nil {
		return err
	}
	if err := insertResultSnapshot(ctx, tx, calculation, r.encryptionKey); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *CalculationRepository) GetByID(ctx context.Context, userID, id string) (domain.Calculation, error) {
	base, err := getCalculationBase(ctx, r.db, userID, id)
	if err != nil {
		return domain.Calculation{}, err
	}
	threshold, err := getThreshold(ctx, r.db, id)
	if err != nil {
		return domain.Calculation{}, err
	}
	employees, err := getPayrollEntries(ctx, r.db, id, r.encryptionKey)
	if err != nil {
		return domain.Calculation{}, err
	}
	monthlyResults, err := getMonthlyResults(ctx, r.db, id, r.encryptionKey)
	if err != nil {
		return domain.Calculation{}, err
	}
	result, engineVersion, err := getResultSnapshot(ctx, r.db, id, r.encryptionKey)
	if err != nil {
		return domain.Calculation{}, err
	}
	result.MonthlyResults = monthlyResults
	return domain.Calculation{
		ID: base.ID,
		Input: domain.CalculationInput{
			OrgName: base.OrgName,
			Title:   base.Title,
			Period: domain.Period{
				Year:    base.PeriodYear,
				Quarter: domain.Quarter(base.PeriodQuarter),
			},
			Threshold: threshold,
			Employees: employees,
		},
		Result:          result,
		EngineVersion:   engineVersion,
		CreatedByUserID: base.CreatedByUserID,
		CreatedAt:       base.CreatedAt,
	}, nil
}

func (r *CalculationRepository) List(ctx context.Context, userID string, filter repository.CalculationListFilter) ([]domain.CalculationSummary, error) {
	var orgName any
	if filter.OrgName != nil {
		orgName = *filter.OrgName
	}
	var year any
	if filter.Year != nil {
		year = *filter.Year
	}
	var quarter any
	if filter.Quarter != nil {
		quarter = int(*filter.Quarter)
	}
	limit := filter.Limit
	if limit < 0 {
		limit = 0
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	rows, err := r.db.QueryContext(ctx, `
	with page as materialized (
	select
		c.id,
		o.name,
		c.title,
		c.period_year,
		c.period_quarter,
		rs.checked_at,
		rs.quarter_average_ciphertext,
		rs.verdict_ciphertext,
		c.created_at
	from calculations c
	join organizations o on o.id = c.organization_id
	join organization_members om on om.organization_id = o.id
	join result_snapshots rs on rs.calculation_id = c.id
	where om.user_id = $1 
	and ($2::text is null or o.normalized_name = lower(regexp_replace(btrim($2::text), '\s+', ' ', 'g')))
	and ($3::integer is null or c.period_year = $3)
	and ($4::integer is null or c.period_quarter = $4)
	order by c.created_at desc, c.id desc
	limit nullif($5, 0) offset $6
	)
	select
		id, name, title, period_year, period_quarter, checked_at,
		pgp_sym_decrypt(quarter_average_ciphertext, $7)::bigint,
		pgp_sym_decrypt(verdict_ciphertext, $7),
		created_at
	from page
	order by created_at desc, id desc
	`, userID, orgName, year, quarter, limit, offset, r.encryptionKey)
	if err != nil {
		return nil, errDecryptCalculationData
	}
	defer rows.Close()
	summaries := make([]domain.CalculationSummary, 0)
	for rows.Next() {
		var id string
		var orgName string
		var title string
		var year int
		var quarter int
		var checkedAt time.Time
		var average int64
		var verdict string
		var createdAt time.Time
		if err := rows.Scan(
			&id,
			&orgName,
			&title,
			&year,
			&quarter,
			&checkedAt,
			&average,
			&verdict,
			&createdAt,
		); err != nil {
			return nil, errDecryptCalculationData
		}
		summary := domain.CalculationSummary{
			ID:             id,
			OrgName:        orgName,
			Title:          title,
			Period:         domain.Period{Year: year, Quarter: domain.Quarter(quarter)},
			CheckedAt:      checkedAt,
			QuarterAverage: domain.NewMoney(average),
			Verdict:        domain.Verdict(verdict),
			CreatedAt:      createdAt,
		}
		summaries = append(summaries, summary)
	}
	if err := rows.Err(); err != nil {
		return nil, errDecryptCalculationData
	}
	return summaries, nil
}

func validateEncryptedCalculation(calculation domain.Calculation) error {
	result := calculation.Result
	if !allNonNegative(
		result.Quarter.MonthlyAverageSum.GetCents(),
		result.Quarter.Average.GetCents(),
		result.Recommendation.EstimatedQuarterFundChange.GetCents(),
	) {
		return errInvalidEncryptedData
	}
	for _, employee := range calculation.Input.Employees {
		for _, month := range employee.Months {
			if !allNonNegative(month.Total140.GetCents(), month.GPH160.GetCents()) {
				return errInvalidEncryptedData
			}
		}
	}
	for _, month := range result.MonthlyResults {
		if month.Headcount < 0 || !allNonNegative(month.LaborPaymentSum.GetCents(), month.Average.GetCents()) {
			return errInvalidEncryptedData
		}
	}
	if result.Threshold.Value != nil && !result.Threshold.Value.IsPositive() {
		return errInvalidEncryptedData
	}
	if percent := result.Comparison.DeltaPercent; percent != nil {
		value, err := strconv.ParseFloat(*percent, 64)
		if err != nil || math.IsInf(value, 0) || math.IsNaN(value) {
			return errInvalidEncryptedData
		}
	}
	if !validVerdict(result.Comparison.Verdict) || !validRecommendationType(result.Recommendation.Type) {
		return errInvalidEncryptedData
	}
	return nil
}

func allNonNegative(values ...int64) bool {
	for _, value := range values {
		if value < 0 {
			return false
		}
	}
	return true
}

func validVerdict(verdict domain.Verdict) bool {
	return verdict == domain.VerdictPass || verdict == domain.VerdictFail || verdict == domain.VerdictNoThreshold
}

func validRecommendationType(recommendation domain.RecommendationType) bool {
	return recommendation == domain.RecommendationCanDecrease || recommendation == domain.RecommendationNeedIncrease || recommendation == domain.RecommendationNoThreshold
}

func (r *CalculationRepository) Delete(ctx context.Context, userID, id string) error {
	result, err := r.db.ExecContext(ctx, `delete from calculations c 
	using organization_members om
	where c.id = $1
	and om.organization_id = c.organization_id
	and om.user_id = $2
	`, id, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrCalculationNotFound
	}

	return nil
}
