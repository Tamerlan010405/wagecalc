package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/domain"
)

func insertPayrollEntries(ctx context.Context, tx *sql.Tx, calculationID string, employees []domain.EmployeePayroll, encryptionKey string) error {
	for _, employee := range employees {
		for monthIndex, month := range employee.Months {
			_, err := tx.ExecContext(ctx, `
				insert into payroll_entries (
					calculation_id,
					employee_index,
					month_index,
					total_140_ciphertext,
					gph_160_ciphertext
				)
				values (
					$1, $2, $3,
					pgp_sym_encrypt($4::text, $6, 'cipher-algo=aes256'),
					pgp_sym_encrypt($5::text, $6, 'cipher-algo=aes256')
				)
			`,
				calculationID,
				employee.EmployeeIndex,
				monthIndex+1,
				month.Total140.GetCents(),
				month.GPH160.GetCents(),
				encryptionKey,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getPayrollEntries(ctx context.Context, db *sql.DB, calculationID, encryptionKey string) ([]domain.EmployeePayroll, error) {
	rows, err := db.QueryContext(ctx, `
		select
			employee_index,
			month_index,
			pgp_sym_decrypt(total_140_ciphertext, $2)::bigint,
			pgp_sym_decrypt(gph_160_ciphertext, $2)::bigint
		from payroll_entries
		where calculation_id = $1
		order by employee_index, month_index
	`, calculationID, encryptionKey)
	if err != nil {
		return nil, errDecryptCalculationData
	}
	defer rows.Close()

	employeesByIndex := make(map[int]*domain.EmployeePayroll)
	employeeIndexes := make([]int, 0)

	for rows.Next() {
		var employeeIndex int
		var monthIndex int
		var total140Cents int64
		var gph160Cents int64
		if err := rows.Scan(
			&employeeIndex,
			&monthIndex,
			&total140Cents,
			&gph160Cents,
		); err != nil {
			return nil, errDecryptCalculationData
		}
		if monthIndex < 1 || monthIndex > domain.CountMonth {
			return nil, fmt.Errorf(
				"payroll entry for calculation %q and employee %d has invalid month index: %d", calculationID, employeeIndex, monthIndex)
		}
		employee, ok := employeesByIndex[employeeIndex]
		if !ok {
			employee = &domain.EmployeePayroll{
				EmployeeIndex: employeeIndex,
			}
			employeesByIndex[employeeIndex] = employee
			employeeIndexes = append(employeeIndexes, employeeIndex)
		}
		employee.Months[monthIndex-1] = domain.MonthPayroll{
			Total140: domain.NewMoney(total140Cents),
			GPH160:   domain.NewMoney(gph160Cents),
		}
	}
	if err := rows.Err(); err != nil {
		return nil, errDecryptCalculationData
	}
	sort.Ints(employeeIndexes)
	employees := make([]domain.EmployeePayroll, 0, len(employeeIndexes))
	for _, employeeIndex := range employeeIndexes {
		employees = append(employees, *employeesByIndex[employeeIndex])
	}
	return employees, nil
}
