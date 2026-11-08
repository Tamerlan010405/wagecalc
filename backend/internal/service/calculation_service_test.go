package service

import (
	"context"
	"testing"
	"time"

	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/domain"
	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/repository"
)

type fakeIDGenerator struct {
	id string
}

func (g fakeIDGenerator) NewID() string {
	return g.id
}

type fakeCalculationRepository struct {
	created domain.Calculation

	createCalled bool
	listCalled   bool
	deleteCalled bool

	getByIDUserID string
	getByIDArg    string
	getByIDResult domain.Calculation
	getByIDErr    error

	listUserID string
	listFilter repository.CalculationListFilter
	listResult []domain.CalculationSummary
	listErr    error

	deleteUserID string
	deleteArg    string
	deleteErr    error
}

func (r *fakeCalculationRepository) Create(ctx context.Context, calculation domain.Calculation) error {
	r.createCalled = true
	r.created = calculation
	return nil
}

func (r *fakeCalculationRepository) GetByID(ctx context.Context, userID string, id string) (domain.Calculation, error) {
	r.getByIDUserID = userID
	r.getByIDArg = id
	if r.getByIDErr != nil {
		return domain.Calculation{}, r.getByIDErr
	}
	return r.getByIDResult, nil
}

func (r *fakeCalculationRepository) List(ctx context.Context, userID string, filter repository.CalculationListFilter) ([]domain.CalculationSummary, error) {
	r.listCalled = true
	r.listUserID = userID
	r.listFilter = filter
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.listResult, nil
}

func (r *fakeCalculationRepository) Delete(ctx context.Context, userID string, id string) error {
	r.deleteCalled = true
	r.deleteUserID = userID
	r.deleteArg = id
	return r.deleteErr
}

func TestCalculationService_CreateCalculation(t *testing.T) {
	fixedTime := time.Date(2025, 7, 6, 12, 0, 0, 0, time.UTC)
	repo := &fakeCalculationRepository{}
	service := NewCalculationService(repo, fakeIDGenerator{id: "calculation-id"})
	service.now = func() time.Time {
		return fixedTime
	}
	input := domain.CalculationInput{
		Period: domain.Period{Year: 2025, Quarter: domain.Quarter3},
		Threshold: domain.Threshold{
			ValueRF: moneyPtr(6500000),
		},
		Employees: employeesFromMonthlyPayments([][3]int64{
			{6500000, 6500000, 6500000},
		}),
	}
	calculation, err := service.CreateCalculation(context.Background(), "user-id", input)
	if err != nil {
		t.Fatalf("CreateCalculation() error = %v", err)
	}
	if calculation.ID != "calculation-id" {
		t.Fatalf("ID = %s, want calculation-id", calculation.ID)
	}
	if calculation.EngineVersion != domain.EngineVersion {
		t.Fatalf("EngineVersion = %s, want %s", calculation.EngineVersion, domain.EngineVersion)
	}
	if !calculation.CreatedAt.Equal(fixedTime) {
		t.Fatalf("CreatedAt = %s, want %s", calculation.CreatedAt, fixedTime)
	}
	if !calculation.Result.CheckedAt.Equal(fixedTime) {
		t.Fatalf("Result.CheckedAt = %s, want %s", calculation.Result.CheckedAt, fixedTime)
	}
	if calculation.Result.Comparison.Verdict != domain.VerdictPass {
		t.Fatalf("Comparison.Verdict = %s, want %s", calculation.Result.Comparison.Verdict, domain.VerdictPass)
	}
	if !repo.createCalled {
		t.Fatal("repository.Create was not called")
	}
	if repo.created.ID != calculation.ID {
		t.Fatalf("saved ID = %s, want %s", repo.created.ID, calculation.ID)
	}
	if !repo.created.Result.CheckedAt.Equal(fixedTime) {
		t.Fatalf("saved Result.CheckedAt = %s, want %s", repo.created.Result.CheckedAt, fixedTime)
	}
	if repo.created.CreatedByUserID != "user-id" {
		t.Fatalf("CreatedByUserID = %s, want user-id", repo.created.CreatedByUserID)
	}
}

func TestCalculationService_GetCalculation(t *testing.T) {
	expected := domain.Calculation{
		ID: "calculation-id",
	}
	repo := &fakeCalculationRepository{
		getByIDResult: expected,
	}
	service := NewCalculationService(repo, fakeIDGenerator{id: "unused-id"})
	calculation, err := service.GetCalculation(context.Background(), "user-id", "calculation-id")
	if err != nil {
		t.Fatalf("GetCalculation() error = %v", err)
	}
	if repo.getByIDArg != "calculation-id" {
		t.Fatalf("GetByID id = %s, want calculation-id", repo.getByIDArg)
	}
	if calculation.ID != expected.ID {
		t.Fatalf("ID = %s, want %s", calculation.ID, expected.ID)
	}
	if repo.getByIDUserID != "user-id" {
		t.Fatalf("GetByID userID = %s, want user-id", repo.getByIDUserID)
	}
}

func TestCalculationService_ListCalculations(t *testing.T) {
	quarter := domain.Quarter3
	filter := repository.CalculationListFilter{
		Quarter: &quarter,
		Limit:   20,
		Offset:  0,
	}
	expected := []domain.CalculationSummary{
		{
			ID:             "calculation-id",
			Period:         domain.Period{Year: 2025, Quarter: domain.Quarter3},
			QuarterAverage: domain.NewMoney(6500000),
			Verdict:        domain.VerdictPass,
		},
	}
	repo := &fakeCalculationRepository{
		listResult: expected,
	}
	service := NewCalculationService(repo, fakeIDGenerator{id: "unused-id"})
	items, err := service.ListCalculations(context.Background(), "user-id", filter)
	if err != nil {
		t.Fatalf("ListCalculations() error = %v", err)
	}
	if !repo.listCalled {
		t.Fatal("repository.List was not called")
	}
	if repo.listFilter.Limit != filter.Limit {
		t.Fatalf("List filter Limit = %d, want %d", repo.listFilter.Limit, filter.Limit)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].ID != expected[0].ID {
		t.Fatalf("items[0].ID = %s, want %s", items[0].ID, expected[0].ID)
	}
	if repo.listUserID != "user-id" {
		t.Fatalf("List userID = %s, want user-id", repo.listUserID)
	}
}

func TestCalculationService_DeleteCalculation(t *testing.T) {
	repo := &fakeCalculationRepository{}
	service := NewCalculationService(repo, fakeIDGenerator{id: "unused-id"})
	err := service.DeleteCalculation(context.Background(), "user-id", "calculation-id")
	if err != nil {
		t.Fatalf("DeleteCalculation() error = %v", err)
	}
	if !repo.deleteCalled {
		t.Fatal("repository.Delete was not called")
	}
	if repo.deleteArg != "calculation-id" {
		t.Fatalf("Delete id = %s, want calculation-id", repo.deleteArg)
	}
	if repo.deleteUserID != "user-id" {
		t.Fatalf("Delete userID = %s, want user-id", repo.deleteUserID)
	}
}
