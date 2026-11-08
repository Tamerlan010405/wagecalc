package service

import (
	"errors"
	"testing"
	"time"

	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/domain"
)

func TestCalculationEvaluator_Evaluate_FNSTestVector(t *testing.T) {
	fixedTime := time.Date(2025, 7, 6, 12, 0, 0, 0, time.UTC)
	evaluator := &CalculationEvaluator{
		now: func() time.Time {
			return fixedTime
		},
	}
	payments := repeatedPayments(11, [3]int64{6473300, 6489500, 6507800})
	payments = append(payments, [3]int64{0, 6489500, 6507800})
	input := domain.CalculationInput{
		Period: domain.Period{Year: 2025, Quarter: domain.Quarter3},
		Threshold: domain.Threshold{
			ValueRF: moneyPtr(6132867),
		},
		Employees: employeesFromMonthlyPayments(payments),
	}
	result, err := evaluator.Evaluate(input)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if result.ThresholdPeriod.Year != 2025 {
		t.Fatalf("ThresholdPeriod.Year = %d, want 2025", result.ThresholdPeriod.Year)
	}
	if result.ThresholdPeriod.Quarter != domain.Quarter2 {
		t.Fatalf("ThresholdPeriod.Quarter = %d, want %d", result.ThresholdPeriod.Quarter, domain.Quarter2)
	}
	if len(result.MonthlyResults) != 3 {
		t.Fatalf("len(MonthlyResults) = %d, want 3", len(result.MonthlyResults))
	}
	assertMonthlyResult(t, result.MonthlyResults[0], 1, 71206300, 11, 6473300, true)
	assertMonthlyResult(t, result.MonthlyResults[1], 2, 77874000, 12, 6489500, true)
	assertMonthlyResult(t, result.MonthlyResults[2], 3, 78093600, 12, 6507800, true)
	if result.Quarter.MonthsWithPayments != 3 {
		t.Fatalf("Quarter.MonthsWithPayments = %d, want 3", result.Quarter.MonthsWithPayments)
	}
	assertMoney(t, result.Quarter.MonthlyAverageSum, 19470600, "Quarter.MonthlyAverageSum")
	assertMoney(t, result.Quarter.Average, 6490200, "Quarter.Average")
	if result.Threshold.Source != domain.ThresholdSourceRF {
		t.Fatalf("Threshold.Source = %s, want %s", result.Threshold.Source, domain.ThresholdSourceRF)
	}
	assertMoneyPtr(t, result.Threshold.Value, 6132867, "Threshold.Value")

	if result.Comparison.Verdict != domain.VerdictPass {
		t.Fatalf("Comparison.Verdict = %s, want %s", result.Comparison.Verdict, domain.VerdictPass)
	}
	assertMoneyPtr(t, result.Comparison.Delta, 357333, "Comparison.Delta")
	if result.Recommendation.Type != domain.RecommendationCanDecrease {
		t.Fatalf("Recommendation.Type = %s, want %s", result.Recommendation.Type, domain.RecommendationCanDecrease)
	}
	assertMoney(t, result.Recommendation.SalaryChangePerEmployeePerMonth, 357333, "Recommendation.SalaryChangePerEmployeePerMonth")
	assertMoney(t, result.Recommendation.EstimatedQuarterFundChange, 12506655, "Recommendation.EstimatedQuarterFundChange")
	if len(result.Warnings) != 0 {
		t.Fatalf("len(Warnings) = %d, want 0", len(result.Warnings))
	}
	if !result.CheckedAt.Equal(fixedTime) {
		t.Fatalf("CheckedAt = %s, want %s", result.CheckedAt, fixedTime)
	}
}

func TestCalculationEvaluator_Evaluate_EmptyMonth(t *testing.T) {
	evaluator := NewCalculationEvaluator()
	input := domain.CalculationInput{
		Period:    domain.Period{Year: 2025, Quarter: domain.Quarter3},
		Employees: employeesFromMonthlyPayments(repeatedPayments(10, [3]int64{6000000, 0, 6600000})),
	}
	result, err := evaluator.Evaluate(input)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	assertMonthlyResult(t, result.MonthlyResults[0], 1, 60000000, 10, 6000000, true)
	assertMonthlyResult(t, result.MonthlyResults[1], 2, 0, 0, 0, false)
	assertMonthlyResult(t, result.MonthlyResults[2], 3, 66000000, 10, 6600000, true)
	if result.Quarter.MonthsWithPayments != 2 {
		t.Fatalf("Quarter.MonthsWithPayments = %d, want 2", result.Quarter.MonthsWithPayments)
	}
	assertMoney(t, result.Quarter.MonthlyAverageSum, 12600000, "Quarter.MonthlyAverageSum")
	assertMoney(t, result.Quarter.Average, 6300000, "Quarter.Average")
}

func TestCalculationEvaluator_Evaluate_AllZeroFailsWithPositiveThreshold(t *testing.T) {
	evaluator := NewCalculationEvaluator()
	input := domain.CalculationInput{
		Period: domain.Period{Year: 2025, Quarter: domain.Quarter3},
		Threshold: domain.Threshold{
			ValueRF: moneyPtr(1000000),
		},
		Employees: employeesFromMonthlyPayments([][3]int64{
			{0, 0, 0},
		}),
	}
	result, err := evaluator.Evaluate(input)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	assertMoney(t, result.Quarter.Average, 0, "Quarter.Average")
	if result.Comparison.Verdict != domain.VerdictFail {
		t.Fatalf("Comparison.Verdict = %s, want %s", result.Comparison.Verdict, domain.VerdictFail)
	}
	if result.Recommendation.Type != domain.RecommendationNeedIncrease {
		t.Fatalf("Recommendation.Type = %s, want %s", result.Recommendation.Type, domain.RecommendationNeedIncrease)
	}
}

func TestCalculationEvaluator_Evaluate_GPHExcludedFromHeadcount(t *testing.T) {
	evaluator := NewCalculationEvaluator()
	input := domain.CalculationInput{
		Period: domain.Period{Year: 2025, Quarter: domain.Quarter3},
		Employees: []domain.EmployeePayroll{
			employeeWithPayroll(1, [3]domain.MonthPayroll{
				{Total140: domain.NewMoney(10000000), GPH160: domain.NewMoney(10000000)},
				{Total140: domain.NewMoney(0), GPH160: domain.NewMoney(0)},
				{Total140: domain.NewMoney(0), GPH160: domain.NewMoney(0)},
			}),
		},
	}
	result, err := evaluator.Evaluate(input)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	assertMonthlyResult(t, result.MonthlyResults[0], 1, 0, 0, 0, false)
	if len(result.Warnings) != 0 {
		t.Fatalf("len(Warnings) = %d, want 0", len(result.Warnings))
	}
}

func TestCalculationEvaluator_Evaluate_UsesLowerRegionThreshold(t *testing.T) {
	evaluator := NewCalculationEvaluator()
	input := domain.CalculationInput{
		Period: domain.Period{Year: 2025, Quarter: domain.Quarter3},
		Threshold: domain.Threshold{
			ValueRF:     moneyPtr(7000000),
			ValueRegion: moneyPtr(6100000),
		},
		Employees: employeesFromMonthlyPayments([][3]int64{
			{6500000, 6500000, 6500000},
		}),
	}
	result, err := evaluator.Evaluate(input)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if result.Threshold.Source != domain.ThresholdSourceRegion {
		t.Fatalf("Threshold.Source = %s, want %s", result.Threshold.Source, domain.ThresholdSourceRegion)
	}
	if result.Comparison.Verdict != domain.VerdictPass {
		t.Fatalf("Comparison.Verdict = %s, want %s", result.Comparison.Verdict, domain.VerdictPass)
	}
}

func TestCalculationEvaluator_Evaluate_EqualThresholdPasses(t *testing.T) {
	evaluator := NewCalculationEvaluator()
	input := domain.CalculationInput{
		Period: domain.Period{Year: 2025, Quarter: domain.Quarter3},
		Threshold: domain.Threshold{
			ValueRF: moneyPtr(6500000),
		},
		Employees: employeesFromMonthlyPayments([][3]int64{
			{6500000, 6500000, 6500000},
		}),
	}
	result, err := evaluator.Evaluate(input)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if result.Comparison.Verdict != domain.VerdictPass {
		t.Fatalf("Comparison.Verdict = %s, want %s", result.Comparison.Verdict, domain.VerdictPass)
	}
	assertMoneyPtr(t, result.Comparison.Delta, 0, "Comparison.Delta")
}

func TestCalculationEvaluator_Evaluate_RecommendationForDeficit(t *testing.T) {
	evaluator := NewCalculationEvaluator()
	input := domain.CalculationInput{
		Period: domain.Period{Year: 2025, Quarter: domain.Quarter3},
		Threshold: domain.Threshold{
			ValueRF: moneyPtr(6132867),
		},
		Employees: employeesFromMonthlyPayments([][3]int64{
			{6000000, 6000000, 6000000},
		}),
	}
	result, err := evaluator.Evaluate(input)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if result.Comparison.Verdict != domain.VerdictFail {
		t.Fatalf("Comparison.Verdict = %s, want %s", result.Comparison.Verdict, domain.VerdictFail)
	}
	assertMoneyPtr(t, result.Comparison.Delta, -132867, "Comparison.Delta")
	if result.Recommendation.Type != domain.RecommendationNeedIncrease {
		t.Fatalf("Recommendation.Type = %s, want %s", result.Recommendation.Type, domain.RecommendationNeedIncrease)
	}
	assertMoney(t, result.Recommendation.SalaryChangePerEmployeePerMonth, 132867, "Recommendation.SalaryChangePerEmployeePerMonth")
	assertMoney(t, result.Recommendation.EstimatedQuarterFundChange, 398601, "Recommendation.EstimatedQuarterFundChange")
}

func TestCalculationEvaluator_Evaluate_GPHGreaterThanTotalAddsWarning(t *testing.T) {
	evaluator := NewCalculationEvaluator()
	input := domain.CalculationInput{
		Period: domain.Period{Year: 2025, Quarter: domain.Quarter3},
		Employees: []domain.EmployeePayroll{
			employeeWithPayroll(1, [3]domain.MonthPayroll{
				{Total140: domain.NewMoney(10000), GPH160: domain.NewMoney(20000)},
				{Total140: domain.NewMoney(0), GPH160: domain.NewMoney(0)},
				{Total140: domain.NewMoney(0), GPH160: domain.NewMoney(0)},
			}),
		},
	}
	result, err := evaluator.Evaluate(input)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	assertMonthlyResult(t, result.MonthlyResults[0], 1, 0, 0, 0, false)
	if len(result.Warnings) != 1 {
		t.Fatalf("len(Warnings) = %d, want 1", len(result.Warnings))
	}
	warning := result.Warnings[0]
	if warning.Code != domain.WarningGPHGreaterThanTotal {
		t.Fatalf("Warning.Code = %s, want %s", warning.Code, domain.WarningGPHGreaterThanTotal)
	}
}

func TestCalculationEvaluator_RejectsInvalidThreshold(t *testing.T) {
	tests := []struct {
		name      string
		threshold domain.Threshold
		field     string
	}{
		{"RF", domain.Threshold{ValueRF: moneyPtr(0)}, "threshold.value_rf"},
		{"Region", domain.Threshold{ValueRegion: moneyPtr(0)}, "threshold.value_region"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCalculationEvaluator().Evaluate(domain.CalculationInput{
				Period:    domain.Period{Year: 2025, Quarter: domain.Quarter3},
				Threshold: tt.threshold,
			})

			assertDomainError(t, err, domain.ErrorInvalidThreshold, tt.field)
		})
	}
}

func TestCalculationEvaluator_RejectsDuplicateEmployeeIndex(t *testing.T) {
	employees := employeesFromMonthlyPayments([][3]int64{
		{10000, 10000, 10000},
		{20000, 20000, 20000},
	})
	employees[1].EmployeeIndex = employees[0].EmployeeIndex
	_, err := NewCalculationEvaluator().Evaluate(domain.CalculationInput{
		Period:    domain.Period{Year: 2025, Quarter: domain.Quarter3},
		Employees: employees,
	})
	assertDomainError(t, err, domain.ErrorInvalidEmployeeIndex, "employees[1].employee_index")
}

func assertDomainError(t *testing.T, err error, code domain.ErrorCode, field string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var domainErr domain.DomainError
	if !errors.As(err, &domainErr) {
		t.Fatalf("error type = %T, want domain.DomainError", err)
	}
	if domainErr.Code != code {
		t.Fatalf("error code = %s, want %s", domainErr.Code, code)
	}
	if domainErr.Field != field {
		t.Fatalf("error field = %s, want %s", domainErr.Field, field)
	}
}

func employeesFromMonthlyPayments(payments [][3]int64) []domain.EmployeePayroll {
	employees := make([]domain.EmployeePayroll, 0, len(payments))
	for i, payment := range payments {
		employees = append(employees, domain.EmployeePayroll{
			EmployeeIndex: i + 1,
			Months: [3]domain.MonthPayroll{
				{Total140: domain.NewMoney(payment[0]), GPH160: domain.NewMoney(0)},
				{Total140: domain.NewMoney(payment[1]), GPH160: domain.NewMoney(0)},
				{Total140: domain.NewMoney(payment[2]), GPH160: domain.NewMoney(0)},
			},
		})
	}

	return employees
}

func repeatedPayments(count int, payment [3]int64) [][3]int64 {
	payments := make([][3]int64, count)
	for i := range payments {
		payments[i] = payment
	}
	return payments
}

func assertMonthlyResult(t *testing.T, got domain.MonthlyResult, wantMonthIndex int, wantLaborPaymentSum int64, wantHeadcount int, wantAverage int64, wantIncluded bool) {
	t.Helper()
	if got.MonthIndex != wantMonthIndex {
		t.Fatalf("MonthIndex = %d, want %d", got.MonthIndex, wantMonthIndex)
	}
	assertMoney(t, got.LaborPaymentSum, wantLaborPaymentSum, "LaborPaymentSum")
	if got.Headcount != wantHeadcount {
		t.Fatalf("Headcount = %d, want %d", got.Headcount, wantHeadcount)
	}
	assertMoney(t, got.Average, wantAverage, "Average")
	if got.IncludedInQuarterAverage != wantIncluded {
		t.Fatalf("IncludedInQuarterAverage = %v, want %v", got.IncludedInQuarterAverage, wantIncluded)
	}
}

func assertMoney(t *testing.T, got domain.Money, wantCents int64, field string) {
	t.Helper()
	if got.GetCents() != wantCents {
		t.Fatalf("%s = %s, want %s", field, got.String(), domain.NewMoney(wantCents).String())
	}
}

func assertMoneyPtr(t *testing.T, got *domain.Money, wantCents int64, field string) {
	t.Helper()

	if got == nil {
		t.Fatalf("%s = nil, want %s", field, domain.NewMoney(wantCents).String())
	}

	assertMoney(t, *got, wantCents, field)
}

func moneyPtr(cents int64) *domain.Money {
	money := domain.NewMoney(cents)
	return &money
}

func employeeWithPayroll(employeeIndex int, months [3]domain.MonthPayroll) domain.EmployeePayroll {
	return domain.EmployeePayroll{
		EmployeeIndex: employeeIndex,
		Months:        months,
	}
}
