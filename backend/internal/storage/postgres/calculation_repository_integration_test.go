package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/domain"
	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/repository"
	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/service"
)

const integrationEncryptionKey = "integration-encryption-key-with-32-bytes"

func TestCalculationRepository_Integration_RoundTrip(t *testing.T) {
	tests := map[string]domain.Threshold{
		"with threshold":    {ValueRF: integrationMoneyPtr(6_000_000)},
		"without threshold": {},
	}
	for name, threshold := range tests {
		t.Run(name, func(t *testing.T) {
			repo, db := openIntegrationRepository(t)
			userID := createIntegrationUser(t, db, "roundtrip@example.com")
			want := integrationCalculation(t, threshold)
			want.CreatedByUserID = userID
			if err := repo.Create(context.Background(), want); err != nil {
				t.Fatalf("Create() error = %v", err)
			}

			got, err := repo.GetByID(context.Background(), userID, want.ID)
			if err != nil {
				t.Fatalf("GetByID() error = %v", err)
			}
			want.CreatedAt = want.CreatedAt.UTC()
			want.Result.CheckedAt = want.Result.CheckedAt.UTC()
			got.CreatedAt = got.CreatedAt.UTC()
			got.Result.CheckedAt = got.Result.CheckedAt.UTC()
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("GetByID() result differs:\ngot:  %#v\nwant: %#v", got, want)
			}
		})
	}
}

func TestCalculationRepository_Integration_ListAndAuthorization(t *testing.T) {
	repo, db := openIntegrationRepository(t)
	ownerID := createIntegrationUser(t, db, "owner@example.com")
	otherUserID := createIntegrationUser(t, db, "other@example.com")
	want := integrationCalculation(t, domain.Threshold{ValueRF: integrationMoneyPtr(6_000_000)})
	want.CreatedByUserID = ownerID
	if err := repo.Create(context.Background(), want); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	summaries, err := repo.List(context.Background(), ownerID, repository.CalculationListFilter{Limit: 10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(summaries) != 1 || summaries[0].QuarterAverage != want.Result.Quarter.Average || summaries[0].Verdict != want.Result.Comparison.Verdict {
		t.Fatalf("List() = %#v", summaries)
	}

	if _, err := repo.GetByID(context.Background(), otherUserID, want.ID); !errors.Is(err, repository.ErrCalculationNotFound) {
		t.Fatalf("GetByID() unauthorized error = %v", err)
	}
	summaries, err = repo.List(context.Background(), otherUserID, repository.CalculationListFilter{Limit: 10})
	if err != nil || len(summaries) != 0 {
		t.Fatalf("List() unauthorized = %#v, %v", summaries, err)
	}
	if err := repo.Delete(context.Background(), otherUserID, want.ID); !errors.Is(err, repository.ErrCalculationNotFound) {
		t.Fatalf("Delete() unauthorized error = %v", err)
	}

	broken := integrationCalculation(t, domain.Threshold{})
	broken.ID = "rollback-calculation-id"
	broken.CreatedByUserID = ownerID
	broken.Result.MonthlyResults[1].MonthIndex = broken.Result.MonthlyResults[0].MonthIndex
	if err := repo.Create(context.Background(), broken); err == nil {
		t.Fatal("Create() error = nil for duplicate monthly result")
	}
	var count int
	if err := db.QueryRow("select count(*) from calculations where id = $1", broken.ID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("rollback calculation count = %d, error = %v", count, err)
	}

	if err := repo.Delete(context.Background(), ownerID, want.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if err := db.QueryRow(`
		select (select count(*) from payroll_entries where calculation_id = $1)
		     + (select count(*) from monthly_results where calculation_id = $1)
		     + (select count(*) from result_snapshots where calculation_id = $1)
	`, want.ID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("child row count after delete = %d, error = %v", count, err)
	}
}

func TestCalculationRepository_Integration_CiphertextAndFailures(t *testing.T) {
	repo, db := openIntegrationRepository(t)
	userID := createIntegrationUser(t, db, "crypto@example.com")
	want := integrationCalculation(t, domain.Threshold{})
	want.CreatedByUserID = userID
	if err := repo.Create(context.Background(), want); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	var first, second []byte
	rows, err := db.Query(`
		select total_140_ciphertext
		from payroll_entries
		where calculation_id = $1 and month_index = 1
		order by employee_index
	`, want.ID)
	if err != nil {
		t.Fatalf("select ciphertext: %v", err)
	}
	defer rows.Close()
	for index, target := range []*[]byte{&first, &second} {
		if !rows.Next() || rows.Scan(target) != nil {
			t.Fatalf("ciphertext row %d is unavailable", index)
		}
	}
	if bytes.Equal(first, second) {
		t.Fatal("equal plaintext values produced equal ciphertext")
	}
	if bytes.Contains(first, []byte("10000000")) {
		t.Fatal("ciphertext contains plaintext payroll value")
	}
	wrongKeyRepo := NewCalculationRepository(db, integrationEncryptionKey+"-wrong")
	if _, err := wrongKeyRepo.GetByID(context.Background(), userID, want.ID); !errors.Is(err, errDecryptCalculationData) {
		t.Fatalf("GetByID() with wrong key error = %v", err)
	}
	if _, err := db.Exec(`
		update payroll_entries
		set total_140_ciphertext = set_byte(total_140_ciphertext, 0, get_byte(total_140_ciphertext, 0) # 1)
		where calculation_id = $1 and employee_index = 1 and month_index = 1
	`, want.ID); err != nil {
		t.Fatalf("corrupt ciphertext: %v", err)
	}
	if _, err := repo.GetByID(context.Background(), userID, want.ID); !errors.Is(err, errDecryptCalculationData) {
		t.Fatalf("GetByID() with corrupted ciphertext error = %v", err)
	}
}

func openIntegrationRepository(t *testing.T) (*CalculationRepository, *sql.DB) {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("PingContext() error = %v", err)
	}
	_, err = db.Exec(`
		TRUNCATE result_snapshots, monthly_results, thresholds,
	payroll_entries, calculations, organization_members,
	sessions, auth_identities, users, organizations CASCADE
	`)
	if err != nil {
		t.Fatalf("truncate database: %v", err)
	}

	return NewCalculationRepository(db, integrationEncryptionKey), db
}

func integrationCalculation(t *testing.T, threshold domain.Threshold) domain.Calculation {
	t.Helper()

	input := domain.CalculationInput{
		OrgName:   "Integration Company",
		Title:     "Integration Calculation",
		Period:    domain.Period{Year: 2025, Quarter: domain.Quarter3},
		Threshold: threshold,
		Employees: []domain.EmployeePayroll{
			{
				EmployeeIndex: 1,
				Months: [3]domain.MonthPayroll{
					{Total140: domain.NewMoney(10_000_000)},
					{Total140: domain.NewMoney(10_000), GPH160: domain.NewMoney(20_000)},
					{Total140: domain.NewMoney(9_000_000), GPH160: domain.NewMoney(1_000_000)},
				},
			},
			{
				EmployeeIndex: 2,
				Months: [3]domain.MonthPayroll{
					{Total140: domain.NewMoney(10_000_000)},
					{Total140: domain.NewMoney(5_000_000), GPH160: domain.NewMoney(1_000_000)},
					{Total140: domain.NewMoney(4_000_000), GPH160: domain.NewMoney(500_000)},
				},
			},
		},
	}

	result, err := service.NewCalculationEvaluator().Evaluate(input)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	fixedTime := time.Date(2025, 7, 9, 12, 0, 0, 0, time.UTC)
	result.CheckedAt = fixedTime

	return domain.Calculation{
		ID:            "integration-calculation-id",
		Input:         input,
		Result:        result,
		EngineVersion: domain.EngineVersion,
		CreatedAt:     fixedTime,
	}
}

func integrationMoneyPtr(cents int64) *domain.Money {
	value := domain.NewMoney(cents)
	return &value
}

func createIntegrationUser(t *testing.T, db *sql.DB, email string) string {
	t.Helper()

	var id string
	err := db.QueryRow(`
		insert into users (email, display_name, created_at, updated_at)
		values ($1, 'Integration User', now(), now())
		returning id
	`, email).Scan(&id)
	if err != nil {
		t.Fatalf("create integration user: %v", err)
	}

	return id
}
