package httpapi

import (
	"context"
	"net/http"

	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/domain"
)

type CalculationWriteService interface {
	CreateCalculation(ctx context.Context, userID string, input domain.CalculationInput) (domain.Calculation, error)
}

type CalculationWriteHandler struct {
	service CalculationWriteService
}

func NewCalculationWriteHandler(service CalculationWriteService) *CalculationWriteHandler {
	return &CalculationWriteHandler{
		service: service,
	}
}

func (h *CalculationWriteHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	if h == nil || h.service == nil {
		writeInternalError(w)
		return
	}

	var requestDTO CalculationInputDTO
	if err := decodeJSON(r, &requestDTO); err != nil {
		writeError(w, err)
		return
	}

	input, err := calculationInputFromDTO(requestDTO)
	if err != nil {
		writeError(w, err)
		return
	}

	user, ok := AuthUserFromContext(r.Context())
	if !ok {
		writeError(w, unauthorizedError())
		return
	}

	calculation, err := h.service.CreateCalculation(r.Context(), user.ID, input)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, calculationDetailsToDTO(calculation))
}
