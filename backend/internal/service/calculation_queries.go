package service

import (
	"context"

	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/domain"
	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/repository"
)

func (s *CalculationService) GetCalculation(ctx context.Context, userID, id string) (domain.Calculation, error) {
	return s.repo.GetByID(ctx, userID, id)
}

func (s *CalculationService) ListCalculations(ctx context.Context, userID string, filter repository.CalculationListFilter) ([]domain.CalculationSummary, error) {
	return s.repo.List(ctx, userID, filter)
}
