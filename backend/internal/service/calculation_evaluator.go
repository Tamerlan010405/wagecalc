package service

import (
	"fmt"
	"time"

	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/domain"
)

type CalculationEvaluator struct {
	now func() time.Time
}

type evaluatedMonth struct {
	result         domain.MonthlyResult
	preciseAverage preciseMoney
}

func NewCalculationEvaluator() *CalculationEvaluator {
	return &CalculationEvaluator{
		now: time.Now,
	}
}

func intPtr(value int) *int {
	return &value
}

func validateCalculationInput(input domain.CalculationInput) error {
	if err := input.Period.Validate(); err != nil {
		return err
	}

	if input.Threshold.ValueRF != nil && !input.Threshold.ValueRF.IsPositive() {
		return domain.NewDomainError(
			domain.ErrorInvalidThreshold,
			"threshold.value_rf",
			"threshold must be greater than zero",
		)
	}

	if input.Threshold.ValueRegion != nil && !input.Threshold.ValueRegion.IsPositive() {
		return domain.NewDomainError(
			domain.ErrorInvalidThreshold,
			"threshold.value_region",
			"threshold must be greater than zero",
		)
	}

	seenEmployees := make(map[int]struct{}, len(input.Employees))
	for index, employee := range input.Employees {
		field := fmt.Sprintf("employees[%d].employee_index", index)

		if employee.EmployeeIndex <= 0 {
			return domain.NewDomainError(
				domain.ErrorInvalidEmployeeIndex,
				field,
				"employee index must be greater than zero",
			)
		}

		if _, exists := seenEmployees[employee.EmployeeIndex]; exists {
			return domain.NewDomainError(
				domain.ErrorInvalidEmployeeIndex,
				field,
				"employee index must be unique",
			)
		}

		seenEmployees[employee.EmployeeIndex] = struct{}{}
	}

	return nil
}

func calculateLaborPayment(employee domain.EmployeePayroll, monthID int, monthMeta domain.MonthMeta) (domain.Money, *domain.CalculationWarning) {
	payroll := employee.Months[monthID]
	if payroll.GPH160.GetCents() > payroll.Total140.GetCents() {
		return domain.NewMoney(0), &domain.CalculationWarning{
			Code:          domain.WarningGPHGreaterThanTotal,
			Message:       "GPH amount is greater than total payroll; labor payment was treated as zero",
			EmployeeIndex: intPtr(employee.EmployeeIndex),
			MonthIndex:    intPtr(monthMeta.Index),
		}
	}
	return payroll.Total140.Sub(payroll.GPH160), nil
}

func evaluateMonth(monthID int, monthMeta domain.MonthMeta, employees []domain.EmployeePayroll) (evaluatedMonth, []domain.CalculationWarning, error) {
	warnings := make([]domain.CalculationWarning, 0)
	monthSum := domain.NewMoney(0)
	headcount := 0
	for _, employee := range employees {
		laborPayment, warning := calculateLaborPayment(employee, monthID, monthMeta)
		if warning != nil {
			warnings = append(warnings, *warning)
		}
		if laborPayment.IsPositive() {
			monthSum = monthSum.Add(laborPayment)
			headcount++
		}
	}
	monthAverage := zeroPreciseMoney()
	if headcount > 0 {
		var err error
		monthAverage, err = divideMoneyPrecisely(monthSum, int64(headcount))
		if err != nil {
			return evaluatedMonth{}, nil, err
		}
	}
	return evaluatedMonth{
		result: domain.MonthlyResult{
			MonthIndex:               monthMeta.Index,
			LaborPaymentSum:          monthSum,
			Headcount:                headcount,
			Average:                  monthAverage.RoundMoney(),
			IncludedInQuarterAverage: monthSum.IsPositive(),
		},
		preciseAverage: monthAverage,
	}, warnings, nil
}

func evaluateMonths(input domain.CalculationInput) ([]evaluatedMonth, []domain.CalculationWarning, error) {
	monthsMeta := input.Period.Months()
	months := make([]evaluatedMonth, 0, domain.CountMonth)
	warnings := make([]domain.CalculationWarning, 0)
	for monthID, monthMeta := range monthsMeta {
		month, monthWarnings, err := evaluateMonth(monthID, monthMeta, input.Employees)
		if err != nil {
			return nil, nil, err
		}
		months = append(months, month)
		warnings = append(warnings, monthWarnings...)
	}
	return months, warnings, nil
}

func monthlyResults(months []evaluatedMonth) []domain.MonthlyResult {
	results := make([]domain.MonthlyResult, 0, len(months))
	for _, month := range months {
		results = append(results, month.result)
	}
	return results
}

func formatDeltaPercent(delta domain.Money, threshold domain.Money) *string {
	if threshold.GetCents() <= 0 {
		return nil
	}
	value := float64(delta.GetCents()) / float64(threshold.GetCents()) * 100
	formatted := fmt.Sprintf("%.2f", value)
	return &formatted
}

func buildQuarterResult(months []evaluatedMonth) (domain.QuarterResult, error) {
	averageSum := zeroPreciseMoney()
	monthsWithPayments := 0
	for _, month := range months {
		if !month.result.IncludedInQuarterAverage {
			continue
		}
		averageSum = averageSum.Add(month.preciseAverage)
		monthsWithPayments++
	}
	if monthsWithPayments == 0 {
		return domain.QuarterResult{
			MonthsWithPayments: 0,
			MonthlyAverageSum:  domain.NewMoney(0),
			Average:            domain.NewMoney(0),
		}, nil
	}
	average, err := averageSum.DivInt(int64(monthsWithPayments))
	if err != nil {
		return domain.QuarterResult{}, err
	}
	return domain.QuarterResult{
		MonthsWithPayments: monthsWithPayments,
		MonthlyAverageSum:  averageSum.RoundMoney(),
		Average:            average.RoundMoney(),
	}, nil
}

func buildComparisonResult(companyAverage domain.Money, threshold domain.ResolvedThreshold) domain.ComparisonResult {
	result := domain.ComparisonResult{
		Verdict: domain.VerdictNoThreshold,
	}
	if threshold.Value == nil {
		return result
	}
	thresholdValue := *threshold.Value
	delta := companyAverage.Sub(thresholdValue)
	result.Delta = &delta
	result.DeltaPercent = formatDeltaPercent(delta, thresholdValue)
	if companyAverage.GreaterOrEqual(thresholdValue) {
		result.Verdict = domain.VerdictPass
	} else {
		result.Verdict = domain.VerdictFail
	}
	return result
}

func buildRecommendationResult(comparison domain.ComparisonResult, months []domain.MonthlyResult) domain.RecommendationResult {
	if comparison.Verdict == domain.VerdictNoThreshold || comparison.Delta == nil {
		return domain.RecommendationResult{
			Type: domain.RecommendationNoThreshold,
		}
	}
	change := comparison.Delta.Abs()
	fundChange := domain.NewMoney(0)
	for _, month := range months {
		if !month.IncludedInQuarterAverage {
			continue
		}
		fundChange = fundChange.Add(change.MulInt(int64(month.Headcount)))
	}
	recommendationType := domain.RecommendationCanDecrease
	if comparison.Verdict == domain.VerdictFail {
		recommendationType = domain.RecommendationNeedIncrease
	}
	return domain.RecommendationResult{
		Type:                            recommendationType,
		SalaryChangePerEmployeePerMonth: change,
		EstimatedQuarterFundChange:      fundChange,
	}
}

func (e *CalculationEvaluator) Evaluate(input domain.CalculationInput) (domain.CalculationResult, error) {
	if err := validateCalculationInput(input); err != nil {
		return domain.CalculationResult{}, err
	}
	evaluatedMonths, warnings, err := evaluateMonths(input)
	if err != nil {
		return domain.CalculationResult{}, err
	}
	quarter, err := buildQuarterResult(evaluatedMonths)
	if err != nil {
		return domain.CalculationResult{}, err
	}
	months := monthlyResults(evaluatedMonths)
	threshold := input.Threshold.Resolve()
	comparison := buildComparisonResult(quarter.Average, threshold)
	recommendation := buildRecommendationResult(comparison, months)
	return domain.CalculationResult{
		ThresholdPeriod: input.Period.PreviousQuarter(),
		MonthlyResults:  months,
		Quarter:         quarter,
		Threshold:       threshold,
		Comparison:      comparison,
		Recommendation:  recommendation,
		Warnings:        warnings,
		CheckedAt:       e.now(),
	}, nil
}
