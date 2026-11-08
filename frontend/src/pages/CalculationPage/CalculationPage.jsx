import { Modal, Text, Title } from "@mantine/core";

import { useEffect, useMemo, useState } from "react";
import { Link, useLocation, useNavigate, useParams } from "react-router";

import { EmployeesTable } from "../../components/EmployeesTable/EmployeesTable";
import { CreateReportCard } from "../../components/CreateReportCard/CreateReportCard";
import reportFormClasses from "../../components/CreateReportCard/CreateReportCard.module.css";
import { RosstatPanel } from "../../components/RosstatPanel/RosstatPanel";
import { exportCalculationReport } from "../../utils/exportCalculationReport";
import classes from "./CalculationPage.module.css";

const API_BASE_URL = import.meta.env.VITE_API_URL ?? "";

const MONTHS = [
  "Январь",
  "Февраль",
  "Март",
  "Апрель",
  "Май",
  "Июнь",
  "Июль",
  "Август",
  "Сентябрь",
  "Октябрь",
  "Ноябрь",
  "Декабрь",
];

function createEmployee(employeeIndex){
  return{
    employee_index: employeeIndex,
    months: [
      {
        month_index: 1,
        total_140: "",
        gph_160: "",
      },
      {
        month_index: 2,
        total_140: "",
        gph_160: "",
      },
      {
        month_index: 3,
        total_140: "",
        gph_160: "",
      },
    ],
  };
}

function getQuarterNumber(quarter){
  const quarterNumber = Number(quarter);

  if (!Number.isInteger(quarterNumber) || quarterNumber < 1 || quarterNumber > 4){
    return ["Месяц 1", "Месяц 2", "Месяц 3"];
  }

  const firstMonthIndex = (quarterNumber - 1) * 3;

  return MONTHS.slice(firstMonthIndex, firstMonthIndex + 3);
}

function formatMoney(value){
  if (value === null || value === "" || value === undefined){
    return "-";
  }

  const number = Number(value);

  if (Number.isNaN(number)){
    return value;
  }

  return new Intl.NumberFormat("ru-RU", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(number);
}

function getValidAmountValue(value){
  const normalizedValue = value.trim().replace(/\s/g, "").replace(",", ".");

  if (normalizedValue === "" || /^\d+(\.\d*)?$/.test(normalizedValue)){
    return normalizedValue;
  }

  return null;
}

function hasNonZeroAmount(value){
  return Number(String(value ?? "").trim().replace(",", ".")) !== 0;
}

function hasNonZeroEmployee(employee){
  return employee.months.some((month) =>
    hasNonZeroAmount(month.total_140) || hasNonZeroAmount(month.gph_160)
  );
}

async function getResponseJson(response){
  try{
    return await response.json();
  }catch{
    return {};
  }
}

export function CalculationPage(){
  const location = useLocation();
  const navigate = useNavigate();
  const { calculationId } = useParams();
  const routeState = location.state ?? {};
  const [reportMeta, setReportMeta] = useState({
    title: routeState.title ?? "Новый отчёт",
    org_name: routeState.org_name ?? "",
    period: {
    year: routeState.period?.year ?? routeState.year ?? "",
    quarter: routeState.period?.quarter ?? routeState.quarter,
    },
    region_code: routeState.region_code ?? "",
  });
  const [employees, setEmployees] = useState([
    createEmployee(1),
  ]);
  const [threshold, setThreshold] = useState({
    value_rf: "",
    value_region: "",
    region_code: null,
  });
  const [calculation, setCalculation] = useState(null);
  const [detailsOpened, setDetailsOpened] = useState(false);
  const [editModalOpened, setEditModalOpened] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [isExporting, setIsExporting] = useState(false);
  const [apiError, setApiError] = useState(null);

  const reportTitle = reportMeta.title;
  const orgName = reportMeta.org_name;
  const year = reportMeta.period.year;
  const quarter = reportMeta.period.quarter;
  const regionCode = reportMeta.region_code;
  const monthNames = useMemo(() => getQuarterNumber(quarter), [quarter]);
  const isResultMode = calculation !== null;
  const result = calculation?.result ?? null;
  const monthlyResults = result?.monthly_results ?? [];

  useEffect(() => {
    if (!calculationId){
      return;
    }

    let isMounted = true;

    async function loadCalculation(){
      try{
        setIsLoading(true);
        setApiError(null);

        const response = await fetch(`${API_BASE_URL}/api/v1/calculations/${calculationId}`, {
          credentials: "include",
        });
        const responseBody = await getResponseJson(response);

        if (!response.ok){
          throw new Error(responseBody?.error?.message ?? `Ошибка запроса: ${response.status}`);
        }

        if (!responseBody.input || !responseBody.result){
          throw new Error("Некорректный ответ сервера");
        }

        if (isMounted){
          const input = responseBody.input;

          setCalculation(responseBody);
          setEmployees(input.employees ?? [createEmployee(1)]);
          setThreshold(input.threshold ?? {
            value_rf: "",
            value_region: "",
            region_code: null,
          });
          setReportMeta({
            title: input.title ?? "Сохранённый отчёт",
            org_name: input.org_name ?? "",
            period: {
              year: input.period?.year ?? "",
              quarter: input.period?.quarter,
            },
            region_code: input.threshold?.region_code ?? "",
          });
          setDetailsOpened(false);
        }
      }catch(error){
        if (isMounted){
          setApiError(error.message || "Не удалось загрузить отчёт");
        }
      }finally{
        if (isMounted){
          setIsLoading(false);
        }
      }
    }

    loadCalculation();

    return () => {
      isMounted = false;
    };
  }, [calculationId]);

  function handleEmployeeValueChange(employeeArrayIndex, monthArrayIndex, field, value){
    const validValue = getValidAmountValue(value);

    if (validValue === null){
      return;
    }

    setEmployees((currentEmployees) => currentEmployees.map((employee, currentEmployeeIndex) =>{
      if (currentEmployeeIndex !== employeeArrayIndex){
        return employee;
      }

      return{
        ...employee,
        months: employee.months.map((month, currentMonthIndex) =>{
          if (currentMonthIndex !== monthArrayIndex){
            return month;
          }

          return{
            ...month,
            [field]: validValue,
          };
        }),
      };
    }));
  }

  function addEmployee(){
    setEmployees((currentEmployees) =>[
      ...currentEmployees,
      createEmployee(currentEmployees.length + 1),
    ]);
  }

  function removeEmployee(employeeArrayIndex){
    setEmployees((currentEmployees) =>{
      if (currentEmployees.length === 1){
        return currentEmployees;
      }

      return currentEmployees
        .filter((_, index) => index !== employeeArrayIndex)
        .map((employee, index) =>({
          ...employee,
          employee_index: index + 1,
        }));
    });
  }

  function handleThresholdChange(field, value){
    const validValue = getValidAmountValue(value);

    if (validValue === null){
      return;
    }

    setThreshold((currentThreshold) =>({
      ...currentThreshold,
      [field]: validValue,
    }));
  }

  async function handleCalculate(){
    setApiError(null);

    const employeesToSend = employees
      .filter(hasNonZeroEmployee)
      .map((employee, index) =>({
        ...employee,
        employee_index: index + 1,
      }));

    const requestBody = {
      org_name: orgName,
      title: reportTitle,
      period:{
        year: Number(year),
        quarter: Number(quarter),
      },
      threshold: {
        value_rf: threshold.value_rf,
        value_region: threshold.value_region,
        region_code: threshold.region_code,
      },
      employees: employeesToSend,
    };

    try{
      setIsLoading(true);
      const response = await fetch(
        `${API_BASE_URL}/api/v1/calculations`,
        {
          method: "POST",
          credentials: "include",
          headers:{
            "Content-type": "application/json",
          },
          body: JSON.stringify(requestBody),
        }
      );
      const responseBody = await getResponseJson(response);

      if (!response.ok){
        setApiError(responseBody?.error?.message ?? `Ошибка запроса: ${response.status}`);
        return;
      }

      if (!responseBody.input || !responseBody.result){
        setApiError("Некорректный ответ сервера");
        return;
      }

      setCalculation(responseBody);
      setEmployees(responseBody.input.employees);
      setThreshold(responseBody.input.threshold);
      setReportMeta({
        title: responseBody.input.title ?? reportTitle,
        org_name: responseBody.input.org_name ?? orgName,
        period: {
          year: responseBody.input.period?.year ?? year,
          quarter: responseBody.input.period?.quarter ?? quarter,
        },
        region_code: responseBody.input.threshold?.region_code ?? regionCode,
      });
      setDetailsOpened(false);

      if (responseBody.id){
        navigate(`/calculation/${responseBody.id}`, { replace: true });
      }
    }catch(error) {
      console.error(error);
      setApiError("Не удалось выполнить расчёт. Проверьте соединение с сервером.");
    }finally{
      setIsLoading(false);
    }
  }

  function handleEditNewVersion(){
    setEditModalOpened(true);
  }

  async function handleDownloadReport(){
    try{
      setIsExporting(true);
      setApiError(null);
      await exportCalculationReport(calculation, monthNames);
    }catch(error){
      console.error(error);
      setApiError("Не удалось сформировать Excel-отчёт");
    }finally{
      setIsExporting(false);
    }
  }

  function handleEditReportMeta(values){
    setReportMeta((currentMeta) =>({
      ...currentMeta,
      title: values.title,
      org_name: values.org_name,
      period: {
        ...currentMeta.period,
        year: values.period.year,
        quarter: values.period.quarter,
      },
    }));
    setCalculation(null);
    setDetailsOpened(false);
    setApiError(null);
    setEditModalOpened(false);
  }

  return (
    <section className={classes.Page}>
      <Modal
        opened={editModalOpened}
        onClose={() => setEditModalOpened(false)}
        centered
        size="calc((100vw - 152px) / 3)"
        radius={16}
        overlayProps={{
          backgroundOpacity: 0.15,
          blur: 0,
        }}
        classNames={{
          content: reportFormClasses.ModalContent,
          body: reportFormClasses.ModalBody,
          header: reportFormClasses.ModalHeader
        }}
      >
        <CreateReportCard
          key={`${reportTitle}-${orgName}-${year}-${quarter}`}
          initialValues={reportMeta}
          title="Редактирование отчёта"
          submitLabel="Редактировать отчет"
          onSubmit={handleEditReportMeta}
        />
      </Modal>
      <div className={classes.Breadcrumbs}>
        <Text
          component={Link}
          to="/projects"
          size="sm"
          className={classes.BreadcrumbLink}
        >
          Главная
        </Text>
        <Text size="sm" c="dimmed">
          ›
        </Text>
        <Text size="sm">
          {reportTitle}
        </Text>
      </div>
      <Title order={3} className={classes.Title}>
        Выплаты по сотрудникам
      </Title>
      <div className={classes.Content}>
        <EmployeesTable
          employees={employees}
          monthNames={monthNames}
          monthlyResults={monthlyResults}
          isResultMode={isResultMode}
          onEmployeeValueChange={handleEmployeeValueChange}
          onAddEmployee={addEmployee}
          onRemoveEmployee={removeEmployee}
          onEditNewVersion={handleEditNewVersion}
          onDownloadReport={handleDownloadReport}
          isExporting={isExporting}
          formatMoney={formatMoney}
        />
        <RosstatPanel
          threshold={threshold}
          result={result}
          monthlyResults={monthlyResults}
          monthNames={monthNames}
          detailsOpened={detailsOpened}
          isLoading={isLoading}
          apiError={apiError}
          isResultMode={isResultMode}
          onThresholdChange={handleThresholdChange}
          onCalculate={handleCalculate}
          onDetailsToggle={() =>
            setDetailsOpened(
              (current) => !current
            )
          }
          formatMoney={formatMoney}
          reportTitle={reportTitle}
        />
      </div>
    </section>
  );
}
