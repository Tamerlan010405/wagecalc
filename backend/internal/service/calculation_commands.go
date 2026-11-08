package service

import (
	"context"
	"time"

	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/domain"
	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/repository"
)

type IDGenerator interface {
	NewID() string
}

type CalculationService struct {
	repo        repository.CalculationRepository
	idGenerator IDGenerator
	now         func() time.Time
}

func NewCalculationService(repo repository.CalculationRepository, idGenerator IDGenerator) *CalculationService {
	return &CalculationService{
		repo:        repo,
		idGenerator: idGenerator,
		now:         time.Now,
	}
}

func (s *CalculationService) CreateCalculation(ctx context.Context, userID string, input domain.CalculationInput) (domain.Calculation, error) {
	now := s.now()
	result, err := s.evaluateAt(input, now)
	if err != nil {
		return domain.Calculation{}, err
	}
	calculation := domain.Calculation{
		ID:              s.idGenerator.NewID(),
		Input:           input,
		Result:          result,
		EngineVersion:   domain.EngineVersion,
		CreatedByUserID: userID,
		CreatedAt:       now,
	}
	if err := s.repo.Create(ctx, calculation); err != nil {
		return domain.Calculation{}, err
	}
	return calculation, nil
}

func (s *CalculationService) evaluateAt(input domain.CalculationInput, checkedAt time.Time) (domain.CalculationResult, error) {
	evaluator := &CalculationEvaluator{
		now: func() time.Time {
			return checkedAt
		},
	}
	return evaluator.Evaluate(input)
}

func (s *CalculationService) DeleteCalculation(ctx context.Context, userID string, id string) error {
	return s.repo.Delete(ctx, userID, id)
}
