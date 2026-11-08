package domain

import "time"

const EngineVersion = "salary-criterion-v1"

type MonthPayroll struct {
	Total140 Money
	GPH160   Money
}

type EmployeePayroll struct {
	EmployeeIndex int
	Months        [3]MonthPayroll
}

type CalculationInput struct {
	OrgName   string
	Title     string
	Period    Period
	Threshold Threshold
	Employees []EmployeePayroll
}

type Calculation struct {
	ID              string
	Input           CalculationInput
	Result          CalculationResult
	EngineVersion   string
	CreatedByUserID string
	CreatedAt       time.Time
}

type CalculationSummary struct {
	ID             string
	OrgName        string
	Title          string
	Period         Period
	CheckedAt      time.Time
	QuarterAverage Money
	Verdict        Verdict
	CreatedAt      time.Time
}
