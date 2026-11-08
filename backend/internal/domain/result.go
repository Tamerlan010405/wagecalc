package domain

import "time"

type Verdict string

const (
	VerdictPass        Verdict = "PASS"
	VerdictFail        Verdict = "FAIL"
	VerdictNoThreshold Verdict = "NO_THRESHOLD"
)

type RecommendationType string

const (
	RecommendationCanDecrease  RecommendationType = "CAN_DECREASE"
	RecommendationNeedIncrease RecommendationType = "NEED_INCREASE"
	RecommendationNoThreshold  RecommendationType = "NO_THRESHOLD"
)

type WarningCode string

const (
	WarningGPHGreaterThanTotal WarningCode = "GPH_GREATER_THAN_TOTAL"
)

type CalculationWarning struct {
	Code          WarningCode
	Message       string
	EmployeeIndex *int
	MonthIndex    *int
}

type MonthlyResult struct {
	MonthIndex               int
	LaborPaymentSum          Money
	Headcount                int
	Average                  Money
	IncludedInQuarterAverage bool
}

type QuarterResult struct {
	MonthsWithPayments int
	MonthlyAverageSum  Money
	Average            Money
}

type ComparisonResult struct {
	Delta        *Money
	DeltaPercent *string
	Verdict      Verdict
}

type RecommendationResult struct {
	Type                            RecommendationType
	SalaryChangePerEmployeePerMonth Money
	EstimatedQuarterFundChange      Money
}

type CalculationResult struct {
	ThresholdPeriod Period
	MonthlyResults  []MonthlyResult
	Quarter         QuarterResult
	Threshold       ResolvedThreshold
	Comparison      ComparisonResult
	Recommendation  RecommendationResult
	Warnings        []CalculationWarning
	CheckedAt       time.Time
}
