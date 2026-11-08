import { Button, Paper, Text, TextInput } from "@mantine/core";

import classes from "./EmployeeTable.module.css";

function hasFilledAmount(value){
  return value !== null && value !== undefined && String(value).trim() !== "";
}

function getMonthlyHeadcount(employees, monthArrayIndex){
  return employees.reduce((headcount, employee) =>{
    const month = employee.months[monthArrayIndex];

    if (!month){
      return headcount;
    }

    return hasFilledAmount(month.total_140) || hasFilledAmount(month.gph_160)
      ? headcount + 1
      : headcount;
  }, 0);
}

export function EmployeesTable({
  employees,
  monthNames,
  monthlyResults,
  isResultMode,
  onEmployeeValueChange,
  onAddEmployee,
  onRemoveEmployee,
  onEditNewVersion,
  onDownloadReport,
  isExporting,
  formatMoney,
}){
  const monthlyHeadcounts = monthNames.map((_, monthArrayIndex) =>
    getMonthlyHeadcount(employees, monthArrayIndex)
  );

  return (
    <Paper className={classes.EmployeesCard}>
      <div className={classes.TableHeader}>
        <div className={classes.EmployeeHeader}>
          Сотрудник
        </div>
        {monthNames.map((monthName) => (
          <div
            className={classes.MonthHeader}
            key={monthName}
          >
            <Text fw={600}>
              {monthName}
            </Text>
            <div className={classes.MonthSubheaders}>
              <span>
                Всего
              </span>
              <span>
                ГПХ
              </span>
            </div>
          </div>
        ))}
      </div>

      <div className={classes.EmployeeRows}>
        {employees.map(
          (
            employee,
            employeeArrayIndex
          ) => (
            <div
              className={classes.EmployeeRow}
              key={employee.employee_index}
            >
              <div className={classes.EmployeeNumber}>
                {employee.employee_index}
              </div>

              {employee.months.map(
                (
                  month,
                  monthArrayIndex
                ) => (
                  <div
                    className={classes.MonthInputs}
                    key={month.month_index}
                  >
                    <TextInput
                      value={month.total_140}
                      placeholder="___"
                      inputMode="decimal"
                      readOnly={isResultMode}
                      onChange={(event) =>
                        onEmployeeValueChange(
                          employeeArrayIndex,
                          monthArrayIndex,
                          "total_140",
                          event.currentTarget.value
                        )
                      }
                    />
                    <TextInput
                      value={month.gph_160}
                      placeholder="___"
                      inputMode="decimal"
                      readOnly={isResultMode}
                      onChange={(event) =>
                        onEmployeeValueChange(
                          employeeArrayIndex,
                          monthArrayIndex,
                          "gph_160",
                          event.currentTarget.value
                        )
                      }
                    />
                  </div>
                )
              )}

              {!isResultMode && (
                <button
                  type="button"
                  className={classes.RemoveEmployeeButton}
                  disabled={employees.length === 1}
                  onClick={() =>
                    onRemoveEmployee(
                      employeeArrayIndex
                    )
                  }
                  aria-label={
                    `Удалить сотрудника ${employee.employee_index}`
                  }
                  title={
                    employees.length === 1
                      ? "Нужен хотя бы один сотрудник"
                      : `Удалить сотрудника ${employee.employee_index}`
                  }
                >
                  ×
                </button>
              )}
            </div>
          )
        )}
      </div>

      <div className={classes.TableFooter}>
        <div className={classes.FooterRow}>
          <span>
            Кол-во физлиц
          </span>
          {monthNames.map(
            (monthName, index) => (
              <strong key={monthName}>
                {isResultMode
                  ? monthlyResults[index]?.headcount ?? "–"
                  : monthlyHeadcounts[index]}
              </strong>
            )
          )}
        </div>
        {isResultMode && (
          <div className={classes.FooterRow}>
            <span>
              Сумма выплат
            </span>
            {monthNames.map(
              (monthName, index) => (
                <strong key={monthName}>
                  {monthlyResults[index]
                    ? `${formatMoney(
                        monthlyResults[index]
                          .labor_payment_sum
                      )} ₽`
                    : "–"}
                </strong>
              )
            )}
          </div>
        )}
      </div>

      <div className={classes.TableAction}>
        {!isResultMode ? (
          <Button
            variant="outline"
            className={classes.AddEmployeeButton}
            onClick={onAddEmployee}
          >
            Добавить сотрудника
          </Button>
        ) : (
          <>
            <Button
              variant="outline"
              radius="xl"
              className={classes.EditButton}
              onClick={onEditNewVersion}
            >
              Редактировать новую версию
            </Button>
            <Button
              variant="outline"
              radius="xl"
              className={classes.EditButton}
              onClick={onDownloadReport}
              loading={isExporting}
              rightSection={
                <span className={`material-symbols-outlined ${classes.DownloadIcon}`}>
                  download
                </span>
              }
            >
              Скачать отчёт
            </Button>
          </>
        )}
      </div>
    </Paper>
  );
}
