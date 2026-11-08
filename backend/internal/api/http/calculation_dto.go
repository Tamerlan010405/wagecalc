package httpapi

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/domain"
)

const maxEmployeesPerCalculation = 10000

type PeriodDTO struct {
	Year    int `json:"year"`
	Quarter int `json:"quarter"`
}

type ThresholdDTO struct {
	ValueRF     *string `json:"value_rf"`
	ValueRegion *string `json:"value_region"`
	RegionCode  *string `json:"region_code"`
}

type MonthPayrollDTO struct {
	MonthIndex int    `json:"month_index"`
	Total140   string `json:"total_140"`
	GPH160     string `json:"gph_160"`
}

type EmployeePayrollDTO struct {
	EmployeeIndex int               `json:"employee_index"`
	Months        []MonthPayrollDTO `json:"months"`
}

type CalculationInputDTO struct {
	OrgName   string               `json:"org_name"`
	Title     string               `json:"title"`
	Period    PeriodDTO            `json:"period"`
	Threshold ThresholdDTO         `json:"threshold"`
	Employees []EmployeePayrollDTO `json:"employees"`
}

type MonthlyResultDTO struct {
	MonthIndex               int    `json:"month_index"`
	LaborPaymentSum          string `json:"labor_payment_sum"`
	Headcount                int    `json:"headcount"`
	Average                  string `json:"average"`
	IncludedInQuarterAverage bool   `json:"included_in_quarter_average"`
}

type QuarterResultDTO struct {
	MonthsWithPayments int    `json:"months_with_payments"`
	MonthlyAverageSum  string `json:"monthly_average_sum"`
	Average            string `json:"average"`
}

type ThresholdResultDTO struct {
	Used   *string `json:"used"`
	Source string  `json:"source"`
}

type ComparisonDTO struct {
	Delta        *string `json:"delta"`
	DeltaPercent *string `json:"delta_percent"`
	Verdict      string  `json:"verdict"`
}

type RecommendationDTO struct {
	Type                            string `json:"type"`
	SalaryChangePerEmployeePerMonth string `json:"salary_change_per_employee_per_month"`
	EstimatedQuarterFundChange      string `json:"estimated_quarter_fund_change"`
}

type WarningDTO struct {
	Code          string `json:"code"`
	Message       string `json:"message"`
	EmployeeIndex *int   `json:"employee_index,omitempty"`
	MonthIndex    *int   `json:"month_index,omitempty"`
}

type CalculationResultDTO struct {
	ThresholdPeriod PeriodDTO          `json:"threshold_period"`
	MonthlyResults  []MonthlyResultDTO `json:"monthly_results"`
	Quarter         QuarterResultDTO   `json:"quarter"`
	Threshold       ThresholdResultDTO `json:"threshold"`
	Comparison      ComparisonDTO      `json:"comparison"`
	Recommendation  RecommendationDTO  `json:"recommendation"`
	Warnings        []WarningDTO       `json:"warnings"`
	CheckedAt       string             `json:"checked_at"`
	EngineVersion   string             `json:"engine_version"`
}

type CalculationDetailsDTO struct {
	ID        string               `json:"id"`
	Input     CalculationInputDTO  `json:"input"`
	Result    CalculationResultDTO `json:"result"`
	CreatedAt string               `json:"created_at"`
}

type CalculationSummaryDTO struct {
	ID             string    `json:"id"`
	OrgName        string    `json:"org_name"`
	Title          string    `json:"title"`
	Period         PeriodDTO `json:"period"`
	CheckedAt      string    `json:"checked_at"`
	QuarterAverage string    `json:"quarter_average"`
	Verdict        string    `json:"verdict"`
	CreatedAt      string    `json:"created_at"`
}

type CalculationListDTO struct {
	Items []CalculationSummaryDTO `json:"items"`
}

func calculationInputFromDTO(dto CalculationInputDTO) (domain.CalculationInput, error) {
	orgName := strings.TrimSpace(dto.OrgName)
	if orgName == "" {
		return domain.CalculationInput{}, newValidationError("org_name", "org_name is required")
	}

	title := strings.TrimSpace(dto.Title)
	if title == "" {
		return domain.CalculationInput{}, newValidationError("title", "title is required")
	}

	threshold, err := thresholdFromDTO(dto.Threshold)
	if err != nil {
		return domain.CalculationInput{}, err
	}

	if len(dto.Employees) > maxEmployeesPerCalculation {
		return domain.CalculationInput{}, newValidationError("employees", "employees count must be no more than 10000")
	}
	employees := make([]domain.EmployeePayroll, 0, len(dto.Employees))
	for employeeIndex, employee := range dto.Employees {
		domainEmployee, err := employeePayrollFromDTO(employee, employeeIndex)
		if err != nil {
			return domain.CalculationInput{}, err
		}
		employees = append(employees, domainEmployee)
	}
	return domain.CalculationInput{
		OrgName:   orgName,
		Title:     title,
		Period:    periodFromDTO(dto.Period),
		Threshold: threshold,
		Employees: employees,
	}, nil
}

func periodFromDTO(dto PeriodDTO) domain.Period {
	return domain.Period{
		Year:    dto.Year,
		Quarter: domain.Quarter(dto.Quarter),
	}
}

func periodToDTO(period domain.Period) PeriodDTO {
	return PeriodDTO{
		Year:    period.Year,
		Quarter: int(period.Quarter),
	}
}

func thresholdFromDTO(dto ThresholdDTO) (domain.Threshold, error) {
	valueRF, err := optionalMoneyFromString(dto.ValueRF, "threshold.value_rf")
	if err != nil {
		return domain.Threshold{}, err
	}

	valueRegion, err := optionalMoneyFromString(dto.ValueRegion, "threshold.value_region")
	if err != nil {
		return domain.Threshold{}, err
	}

	var regionCode *string
	if dto.RegionCode != nil {
		trimmedRegionCode := strings.TrimSpace(*dto.RegionCode)
		if trimmedRegionCode != "" {
			regionCode = &trimmedRegionCode
		}
	}

	return domain.Threshold{
		ValueRF:     valueRF,
		ValueRegion: valueRegion,
		RegionCode:  regionCode,
	}, nil
}

func employeePayrollFromDTO(dto EmployeePayrollDTO, employeePosition int) (domain.EmployeePayroll, error) {
	if len(dto.Months) != domain.CountMonth {
		return domain.EmployeePayroll{}, newValidationError(
			fieldPath("employees", employeePosition, "months"),
			"employee must contain exactly 3 months",
		)
	}

	var months [domain.CountMonth]domain.MonthPayroll
	seen := [domain.CountMonth]bool{}
	for monthPosition, month := range dto.Months {
		if month.MonthIndex < 1 || month.MonthIndex > domain.CountMonth {
			return domain.EmployeePayroll{}, newValidationError(
				fieldPath("employees", employeePosition, "months", monthPosition, "month_index"),
				"month_index must be between 1 and 3",
			)
		}

		monthArrayIndex := month.MonthIndex - 1
		if seen[monthArrayIndex] {
			return domain.EmployeePayroll{}, newValidationError(
				fieldPath("employees", employeePosition, "months", monthPosition, "month_index"),
				"month_index must be unique for employee",
			)
		}
		seen[monthArrayIndex] = true

		total140, err := moneyFromString(month.Total140, fieldPath("employees", employeePosition, "months", monthPosition, "total_140"))
		if err != nil {
			return domain.EmployeePayroll{}, err
		}

		gph160, err := moneyFromString(month.GPH160, fieldPath("employees", employeePosition, "months", monthPosition, "gph_160"))
		if err != nil {
			return domain.EmployeePayroll{}, err
		}

		months[monthArrayIndex] = domain.MonthPayroll{
			Total140: total140,
			GPH160:   gph160,
		}
	}

	return domain.EmployeePayroll{
		EmployeeIndex: dto.EmployeeIndex,
		Months:        months,
	}, nil
}

func moneyFromString(value string, field string) (domain.Money, error) {
	if strings.TrimSpace(value) == "" {
		return domain.NewMoney(0), nil
	}

	money, err := domain.ParseMoney(value)
	if err != nil {
		return domain.Money{}, withField(err, field)
	}

	return money, nil
}

func optionalMoneyFromString(value *string, field string) (*domain.Money, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}

	money, err := domain.ParseMoney(*value)
	if err != nil {
		return nil, withField(err, field)
	}

	return &money, nil
}

func withField(err error, field string) error {
	var domainErr domain.DomainError
	if errors.As(err, &domainErr) {
		return domain.NewDomainError(domainErr.Code, field, domainErr.Message)
	}
	return err
}

func calculationDetailsToDTO(calculation domain.Calculation) CalculationDetailsDTO {
	return CalculationDetailsDTO{
		ID:        calculation.ID,
		Input:     calculationInputToDTO(calculation.Input),
		Result:    calculationResultToDTO(calculation.Result, calculation.EngineVersion),
		CreatedAt: timeToString(calculation.CreatedAt),
	}
}

func calculationInputToDTO(input domain.CalculationInput) CalculationInputDTO {
	employees := make([]EmployeePayrollDTO, 0, len(input.Employees))
	for _, employee := range input.Employees {
		employees = append(employees, employeePayrollToDTO(employee))
	}

	return CalculationInputDTO{
		OrgName:   input.OrgName,
		Title:     input.Title,
		Period:    periodToDTO(input.Period),
		Threshold: thresholdToDTO(input.Threshold),
		Employees: employees,
	}
}

func thresholdToDTO(threshold domain.Threshold) ThresholdDTO {
	return ThresholdDTO{
		ValueRF:     moneyPtrToStringPtr(threshold.ValueRF),
		ValueRegion: moneyPtrToStringPtr(threshold.ValueRegion),
		RegionCode:  threshold.RegionCode,
	}
}

func employeePayrollToDTO(employee domain.EmployeePayroll) EmployeePayrollDTO {
	months := make([]MonthPayrollDTO, 0, domain.CountMonth)
	for index, month := range employee.Months {
		months = append(months, MonthPayrollDTO{
			MonthIndex: index + 1,
			Total140:   month.Total140.String(),
			GPH160:     month.GPH160.String(),
		})
	}

	return EmployeePayrollDTO{
		EmployeeIndex: employee.EmployeeIndex,
		Months:        months,
	}
}

func calculationResultToDTO(result domain.CalculationResult, engineVersion string) CalculationResultDTO {
	monthlyResults := make([]MonthlyResultDTO, 0, len(result.MonthlyResults))
	for _, month := range result.MonthlyResults {
		monthlyResults = append(monthlyResults, monthlyResultToDTO(month))
	}

	warnings := make([]WarningDTO, 0, len(result.Warnings))
	for _, warning := range result.Warnings {
		warnings = append(warnings, warningToDTO(warning))
	}

	return CalculationResultDTO{
		ThresholdPeriod: periodToDTO(result.ThresholdPeriod),
		MonthlyResults:  monthlyResults,
		Quarter:         quarterResultToDTO(result.Quarter),
		Threshold:       thresholdResultToDTO(result.Threshold),
		Comparison:      comparisonToDTO(result.Comparison),
		Recommendation:  recommendationToDTO(result.Recommendation),
		Warnings:        warnings,
		CheckedAt:       timeToString(result.CheckedAt),
		EngineVersion:   engineVersion,
	}
}

func monthlyResultToDTO(result domain.MonthlyResult) MonthlyResultDTO {
	return MonthlyResultDTO{
		MonthIndex:               result.MonthIndex,
		LaborPaymentSum:          result.LaborPaymentSum.String(),
		Headcount:                result.Headcount,
		Average:                  result.Average.String(),
		IncludedInQuarterAverage: result.IncludedInQuarterAverage,
	}
}

func quarterResultToDTO(result domain.QuarterResult) QuarterResultDTO {
	return QuarterResultDTO{
		MonthsWithPayments: result.MonthsWithPayments,
		MonthlyAverageSum:  result.MonthlyAverageSum.String(),
		Average:            result.Average.String(),
	}
}

func thresholdResultToDTO(result domain.ResolvedThreshold) ThresholdResultDTO {
	return ThresholdResultDTO{
		Used:   moneyPtrToStringPtr(result.Value),
		Source: string(result.Source),
	}
}

func comparisonToDTO(result domain.ComparisonResult) ComparisonDTO {
	return ComparisonDTO{
		Delta:        moneyPtrToStringPtr(result.Delta),
		DeltaPercent: result.DeltaPercent,
		Verdict:      string(result.Verdict),
	}
}

func recommendationToDTO(result domain.RecommendationResult) RecommendationDTO {
	return RecommendationDTO{
		Type:                            string(result.Type),
		SalaryChangePerEmployeePerMonth: result.SalaryChangePerEmployeePerMonth.String(),
		EstimatedQuarterFundChange:      result.EstimatedQuarterFundChange.String(),
	}
}

func warningToDTO(warning domain.CalculationWarning) WarningDTO {
	return WarningDTO{
		Code:          string(warning.Code),
		Message:       warning.Message,
		EmployeeIndex: warning.EmployeeIndex,
		MonthIndex:    warning.MonthIndex,
	}
}

func calculationSummaryToDTO(summary domain.CalculationSummary) CalculationSummaryDTO {
	return CalculationSummaryDTO{
		ID:             summary.ID,
		OrgName:        summary.OrgName,
		Title:          summary.Title,
		Period:         periodToDTO(summary.Period),
		CheckedAt:      timeToString(summary.CheckedAt),
		QuarterAverage: summary.QuarterAverage.String(),
		Verdict:        string(summary.Verdict),
		CreatedAt:      timeToString(summary.CreatedAt),
	}
}

func calculationListToDTO(summaries []domain.CalculationSummary) CalculationListDTO {
	items := make([]CalculationSummaryDTO, 0, len(summaries))
	for _, summary := range summaries {
		items = append(items, calculationSummaryToDTO(summary))
	}
	return CalculationListDTO{Items: items}
}

func moneyPtrToStringPtr(money *domain.Money) *string {
	if money == nil {
		return nil
	}
	value := money.String()
	return &value
}

func timeToString(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339)
}

func fieldPath(parts ...any) string {
	var builder strings.Builder
	for index, part := range parts {
		switch value := part.(type) {
		case string:
			if index > 0 {
				builder.WriteByte('.')
			}
			builder.WriteString(value)
		case int:
			builder.WriteByte('[')
			builder.WriteString(strconv.Itoa(value))
			builder.WriteByte(']')
		}
	}
	return builder.String()
}
