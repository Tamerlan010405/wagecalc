package repository

import (
	"context"
	"errors"

	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/domain"
)

var ErrCalculationNotFound = errors.New("calculation not found")

type CalculationListFilter struct {
	OrgName *string
	Year    *int
	Quarter *domain.Quarter
	Limit   int
	Offset  int
}

type CalculationRepository interface {
	Create(ctx context.Context, calculation domain.Calculation) error
	GetByID(ctx context.Context, userID, id string) (domain.Calculation, error)
	List(ctx context.Context, userID string, filter CalculationListFilter) ([]domain.CalculationSummary, error)
	Delete(ctx context.Context, userID, id string) error
}
