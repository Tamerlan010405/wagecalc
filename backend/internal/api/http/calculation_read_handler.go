package httpapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/domain"
	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/repository"
)

type CalculationReadService interface {
	ListCalculations(ctx context.Context, userID string, filter repository.CalculationListFilter) ([]domain.CalculationSummary, error)
	GetCalculation(ctx context.Context, userID, id string) (domain.Calculation, error)
	DeleteCalculation(ctx context.Context, userID, id string) error
}

type CalculationReadHandler struct {
	service CalculationReadService
}

func NewCalculationReadHandler(service CalculationReadService) *CalculationReadHandler {
	return &CalculationReadHandler{
		service: service,
	}
}

func (h *CalculationReadHandler) List(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	if h == nil || h.service == nil {
		writeInternalError(w)
		return
	}

	filter, err := calculationListFilterFromRequest(r)
	if err != nil {
		writeError(w, err)
		return
	}
	user, ok := AuthUserFromContext(r.Context())
	if !ok {
		writeError(w, unauthorizedError())
		return
	}
	summaries, err := h.service.ListCalculations(r.Context(), user.ID, filter)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, calculationListToDTO(summaries))
}

func (h *CalculationReadHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	if h == nil || h.service == nil {
		writeInternalError(w)
		return
	}

	id, ok := calculationIDFromRequest(w, r)
	if !ok {
		return
	}
	user, ok := AuthUserFromContext(r.Context())
	if !ok {
		writeError(w, unauthorizedError())
		return
	}
	calculation, err := h.service.GetCalculation(r.Context(), user.ID, id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, calculationDetailsToDTO(calculation))
}

func (h *CalculationReadHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodDelete) {
		return
	}

	if h == nil || h.service == nil {
		writeInternalError(w)
		return
	}

	id, ok := calculationIDFromRequest(w, r)
	if !ok {
		return
	}

	user, ok := AuthUserFromContext(r.Context())
	if !ok {
		writeError(w, unauthorizedError())
		return
	}

	if err := h.service.DeleteCalculation(r.Context(), user.ID, id); err != nil {
		writeError(w, err)
		return
	}

	writeNoContent(w)
}

func calculationIDFromRequest(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeError(w, newAPIError(http.StatusBadRequest, codeInvalidRequest, "calculation id is required", "id", nil))
		return "", false
	}
	return id, true
}

func calculationListFilterFromRequest(r *http.Request) (repository.CalculationListFilter, error) {
	query := r.URL.Query()
	filter := repository.CalculationListFilter{}

	if orgName := strings.TrimSpace(query.Get("org_name")); orgName != "" {
		filter.OrgName = &orgName
	}

	if value := strings.TrimSpace(query.Get("year")); value != "" {
		year, err := strconv.Atoi(value)
		if err != nil {
			return repository.CalculationListFilter{}, newValidationError("year", "year must be an integer")
		}
		filter.Year = &year
	}

	if value := strings.TrimSpace(query.Get("quarter")); value != "" {
		quarterValue, err := strconv.Atoi(value)
		if err != nil {
			return repository.CalculationListFilter{}, newValidationError("quarter", "quarter must be an integer")
		}
		if quarterValue < int(domain.Quarter1) || quarterValue > int(domain.Quarter4) {
			return repository.CalculationListFilter{}, newValidationError("quarter", "quarter must be between 1 and 4")
		}
		quarter := domain.Quarter(quarterValue)
		filter.Quarter = &quarter
	}

	if value := strings.TrimSpace(query.Get("limit")); value != "" {
		limit, err := strconv.Atoi(value)
		if err != nil || limit <= 0 {
			return repository.CalculationListFilter{}, newValidationError("limit", "limit must be a positive integer")
		}
		filter.Limit = limit
	}

	if value := strings.TrimSpace(query.Get("offset")); value != "" {
		offset, err := strconv.Atoi(value)
		if err != nil || offset < 0 {
			return repository.CalculationListFilter{}, newValidationError("offset", "offset must be a non-negative integer")
		}
		filter.Offset = offset
	}

	return filter, nil
}
