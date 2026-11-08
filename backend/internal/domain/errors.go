package domain

type ErrorCode string

const (
	ErrorInvalidPeriod             ErrorCode = "INVALID_PERIOD"
	ErrorInvalidMonthCount         ErrorCode = "INVALID_MONTH_COUNT"
	ErrorNegativeMoney             ErrorCode = "NEGATIVE_MONEY"
	ErrorInvalidEmployeeIndex      ErrorCode = "INVALID_EMPLOYEE_INDEX"
	ErrorInvalidYear               ErrorCode = "INVALID_YEAR"
	ErrorInvalidQuarter            ErrorCode = "INVALID_QUARTER"
	ErrorInvalidMoneyFormat        ErrorCode = "INVALID_MONEY_FORMAT"
	ErrorEmptyMoney                ErrorCode = "MONEY_IS_EMPTY"
	ErrorMoneyTooManyDecimalDigits ErrorCode = "TOO_MANY_DECIMAL_DIGITS"
	ErrorDivisionByZero            ErrorCode = "DIVISION_BY_ZERO"
	ErrorInvalidThreshold          ErrorCode = "INVALID_THRESHOLD"
	ErrorInvalidEmail              ErrorCode = "INVALID_EMAIL"
	ErrorInvalidPassword           ErrorCode = "INVALID_PASSWORD"
	ErrorInvalidPasswordResetCode  ErrorCode = "INVALID_PASSWORD_RESET_CODE"
)

type DomainError struct {
	Code    ErrorCode
	Message string
	Field   string
}

func (e DomainError) Error() string {
	return e.Message
}

func NewDomainError(code ErrorCode, field string, message string) DomainError {
	return DomainError{
		Code:    code,
		Field:   field,
		Message: message,
	}
}
