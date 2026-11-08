package httpapi

import (
	"errors"
	"net/http"

	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/domain"
	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/repository"
)

const (
	codeInvalidRequest   = "INVALID_REQUEST"
	codeValidationError  = "VALIDATION_ERROR"
	codeNotFound         = "NOT_FOUND"
	codeMethodNotAllowed = "METHOD_NOT_ALLOWED"
	codeInternalError    = "INTERNAL_ERROR"
	codeNotImplemented   = "NOT_IMPLEMENTED"
)

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

type apiError struct {
	status  int
	code    string
	message string
	field   string
	err     error
}

func (e *apiError) Error() string {
	if e == nil {
		return ""
	}
	return e.message
}

func newAPIError(status int, code, message, field string, err error) *apiError {
	return &apiError{
		status:  status,
		code:    code,
		message: message,
		field:   field,
		err:     err,
	}
}

func newValidationError(field string, message string) *apiError {
	return newAPIError(http.StatusUnprocessableEntity, codeValidationError, message, field, nil)
}

func writeError(w http.ResponseWriter, err error) {
	apiErr := mapError(err)
	writeJSON(w, apiErr.status, errorResponse{
		Error: errorBody{
			Code:    apiErr.code,
			Message: apiErr.message,
			Field:   apiErr.field,
		},
	})
}

func writeInternalError(w http.ResponseWriter) {
	writeError(w, newAPIError(http.StatusInternalServerError, codeInternalError, "internal server error", "", nil))
}

func mapError(err error) *apiError {
	if err == nil {
		return newAPIError(http.StatusInternalServerError, codeInternalError, "internal server error", "", nil)
	}

	var apiErr *apiError
	if errors.As(err, &apiErr) {
		return apiErr
	}

	var domainErr domain.DomainError
	if errors.As(err, &domainErr) {
		return newAPIError(http.StatusUnprocessableEntity, codeValidationError, domainErr.Message, domainErr.Field, err)
	}

	if errors.Is(err, repository.ErrCalculationNotFound) {
		return newAPIError(http.StatusNotFound, codeNotFound, "calculation not found", "", err)
	}

	return newAPIError(http.StatusInternalServerError, codeInternalError, "internal server error", "", err)
}
