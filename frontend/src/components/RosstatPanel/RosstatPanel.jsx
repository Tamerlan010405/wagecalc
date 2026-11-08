import { Button, Collapse, Group, Paper, Stack, Text, TextInput } from "@mantine/core";
import { useLocation, useNavigate } from "react-router";

import classes from "./RosstatPanel.module.module.css";

function getVerdictData(verdict){
  switch(verdict){
    case "PASS":
      return{
        text: "Критерий выполняется",
        icon: "✓",
        className: classes.PassResult,
      };
    case "FAIL":
      return {
        text: "Критерий не выполняется",
        icon: "×",
        className: classes.FailResult,
      };
    case "NO_THRESHOLD":
      return {
        text: "Планка не задана",
        icon: "?",
        className: classes.NoThresholdResult,
      };
    default:
      return{
        text: "Результат неизвестен",
        icon: "?",
        className: classes.NoThresholdResult,
      };
  }
}

function getThresholdLabel(result){
  if (result?.threshold?.source === "RF"){
    return "По РФ";
  }
  if (result?.threshold?.source === "REGION"){
    return "По региону";
  }
  return "Применённая планка";
}

function getRecommendationText(result, formatMoney){
  const recommendation = result?.recommendation;

  if (!recommendation){
    return null;
  }

  if (recommendation.type === "CAN_DECREASE"){
    return (
      <>
        Запас:{" "}
        <strong>
          {formatMoney(
            recommendation
              .salary_change_per_employee_per_month
          )}{" "}
          ₽/мес
        </strong>
        {" "}на сотрудника
      </>
    );
  }

  if (recommendation.type === "NEED_INCREASE"){
    return (
      <>
        Нужно увеличить на:{" "}
        <strong>
          {formatMoney(
            recommendation
              .salary_change_per_employee_per_month
          )}{" "}
          ₽/мес
        </strong>
        {" "}на сотрудника
      </>
    );
  }

  return "Планка для сравнения не задана";
}

export function RosstatPanel({
  threshold,
  result,
  monthlyResults,
  monthNames,
  detailsOpened,
  isLoading,
  apiError,
  isResultMode,
  onThresholdChange,
  onCalculate,
  onDetailsToggle,
  formatMoney,
  reportTitle,
}){

  const navigate = useNavigate();
  const location = useLocation();

  function handleClick(){
    navigate("/manual", {
      state: {
        reportTitle,
        reportPath: `${location.pathname}${location.search}`,
      },
    });
  }

  const verdictData = result
    ? getVerdictData(result.comparison.verdict)
    : null;
  const thresholdLabel = getThresholdLabel(result);
  const hasRfValue = String(threshold.value_rf ?? "").trim() !== "";
  const hasRegionValue = String(threshold.value_region ?? "").trim() !== "";
  const rfValue = Number(String(threshold.value_rf ?? "").replace(",", "."));
  const regionValue = Number(String(threshold.value_region ?? "").replace(",", "."));
  const isRfLover = hasRfValue && (!hasRegionValue || rfValue < regionValue);
  const isRegionLover = hasRegionValue && (!hasRfValue || regionValue < rfValue);

  return (
    <Stack
      className={`${classes.Sidebar} ${detailsOpened ? classes.SideBarExpanded : ""}`}
      gap={16}
    >
      <Paper className={classes.ThresholdCard}>
        <Text fw={700}>
          Планка Росстата
        </Text>
        {!isResultMode ? (
          <>
            <Group
              grow
              gap={8}
              className={classes.ThresholdInputs}
            >
              <TextInput
                className = {isRfLover ? classes.ChoiseBar : ""}
                label="По РФ"
                placeholder="-"
                inputMode="decimal"
                value={threshold.value_rf}
                onChange={(event) =>
                  onThresholdChange(
                    "value_rf",
                    event.currentTarget.value
                  )
                }
              />
              <TextInput
                className = {isRegionLover ? classes.ChoiseBar : ""}
                label="По региону"
                placeholder="-"
                inputMode="decimal"
                value={threshold.value_region}
                onChange={(event) =>
                  onThresholdChange(
                    "value_region",
                    event.currentTarget.value
                  )
                }
              />
            </Group>
            <Text
              size="sm"
              className={classes.ThresholdHint}
            >
              Принимается меньшая планка
            </Text>
            <div className={classes.CalculateButtonWrapper}>
              <Button
                radius="xl"
                loading={isLoading}
                onClick={onCalculate}
              >
                Рассчитать
              </Button>
            </div>
          </>
        ) : (
          <Text
            fw={600}
            c="dimmed"
            mt={6}
          >
            {thresholdLabel}:{" "}
            {formatMoney(
              result.threshold.used
            )}{" "}
            ₽
          </Text>
        )}
      </Paper>

      {isResultMode && (
        <Paper className={classes.ResultCard}>
          <Group justify="space-between" className={classes.ResultHeader}>
            <Text
              fw={700}
            >
              Результат
            </Text>
            <Button className={classes.ButtonCalculationMethod} onClick={handleClick} variant="transparent">
              Методика и допущения
            </Button>
          </Group>
          <div
            className={
              `${classes.ResultSummary} ${verdictData.className}`
            }
          >
            <div className={classes.VerdictRow}>
              <span className={`material-symbols-outlined ${classes.VerdictIcon}`}>
                {result.comparison.verdict === "PASS" ? "check" : "close"}
              </span>
              <Text fw={600}>
                {verdictData.text}
              </Text>
            </div>
            <div className={classes.AverageValue}>
              {formatMoney(
                result.quarter.average
              )}{" "}
              ₽
              <span>
                /мес
              </span>
            </div>
            <div className={classes.Recommendation}>
              {getRecommendationText(result, formatMoney)}
            </div>
          </div>
          <Collapse expanded={detailsOpened}>
            <div className={classes.Steps}>
              <div className={classes.Step}>
                <div className={classes.StepDot} />
                <div className={classes.StepContent}>
                  <div className={classes.tepHeader}>
                    <Text fw={600}>
                      Выплаты по ТД
                    </Text>
                    <Text
                      fw={700}
                      className={`${classes.StepValue} ${
                    result?.comparison.verdict === "PASS"?classes.PassResultValue:classes.FailResultValue
                  }`}
                    >
                      140 – 160
                    </Text>
                  </div>
                  <Text className={classes.TextPoint}>
                    п. 5, абз. 2 письма ФНС
                  </Text>
                  <Text
                    size="xs"
                    c="dimmed"
                  >
                    По каждому физлицу за месяц: сумма выплат минус ГПХ
                  </Text>
                </div>
              </div>

              <div className={classes.Step}>
                <div className={classes.StepDot} />

                <div className={classes.StepContent}>
                  <div className={classes.StepHeader}>
                    <Text fw={600}>
                      Средняя за месяц
                    </Text>

                    <Text
                      fw={700}
                      className={`${classes.StepValue} ${
                    result?.comparison.verdict === "PASS"?classes.PassResultValue:classes.FailResultValue
                  }`}
                    >
                      Σ ÷ физлица
                    </Text>
                  </div>

                  <Text size="xs">
                    п. 5, абз. 2
                  </Text>

                  {monthlyResults.map(
                    (
                      monthResult,
                      index
                    ) => (
                      <Text
                        size="xs"
                        c="dimmed"
                        key={monthResult.month_index}
                      >
                        {monthNames[index]}:{" "}
                        {formatMoney(
                          monthResult
                            .labor_payment_sum
                        )}
                        {" "}÷{" "}
                        {monthResult.headcount}
                        {" "}={" "}
                        {formatMoney(
                          monthResult.average
                        )}
                      </Text>
                    )
                  )}
                </div>
              </div>

              <div className={classes.Step}>
                <div className={classes.StepDot} />

                <div className={classes.StepContent}>
                  <div className={classes.StepHeader}>
                    <Text fw={600}>
                      Среднемесячная за квартал
                    </Text>

                    <Text
                      fw={700}
                      className={`${classes.StepValue} ${
                    result?.comparison.verdict === "PASS"?classes.PassResultValue:classes.FailResultValue
                  }`}
                    >
                      {formatMoney(
                        result.quarter.average
                      )}{" "}
                      ₽
                    </Text>
                  </div>
                  <Text size="xs">
                    пп. 6 п. 4
                  </Text>
                  <Text
                    size="xs"
                    c="dimmed"
                  >
                    Сумма средних за месяцы с выплатами:{" "}
                    {formatMoney(
                      result.quarter
                        .monthly_average_sum
                    )}
                    {" "}÷{" "}
                    {
                      result.quarter
                        .months_with_payments
                    }
                    {" "}={" "}
                    {formatMoney(
                      result.quarter.average
                    )}
                  </Text>
                </div>
              </div>
              <div className={classes.Step}>
                <div className={classes.StepDot} />
                <div className={classes.StepContent}>
                  <div className={classes.StepHeader}>
                    <Text fw={600}>
                      Сравнение с планкой
                    </Text>
                    <Text
                      fw={700}
                      className={`${classes.StepValue} ${
                    result?.comparison.verdict === "PASS"?classes.PassResultValue:classes.FailResultValue
                  }`}
                    >
                      {formatMoney(
                        result.threshold.used
                      )}{" "}
                      ₽
                    </Text>
                  </div>
                  <Text size="xs">
                    пп. 6 п. 4, Росстат
                  </Text>
                  {result.comparison.verdict !==
                    "NO_THRESHOLD" && (
                    <Text
                      size="xs"
                      c="dimmed"
                    >
                      Разница:{" "}
                      {formatMoney(
                        result.comparison.delta
                      )}{" "}
                      ₽,{" "}
                      {
                        result.comparison
                          .delta_percent
                      }
                      %
                    </Text>
                  )}
                </div>
              </div>
            </div>
          </Collapse>
          <button
            type="button"
            className={classes.DetailsButton}
            onClick={onDetailsToggle}
          >
            {detailsOpened
              ? "Скрыть расчёт"
              : "Подробный расчёт"}
            <span className={`material-symbols-outlined ${classes.Arrow}`}>
              {`${detailsOpened? "keyboard_arrow_up" : "keyboard_arrow_down"
                }`}
            </span>
          </button>
        </Paper>
      )}

      {apiError && (
        <Text
          className={classes.ApiError}
          c="red"
          size="sm"
        >
          {apiError}
        </Text>
      )}
    </Stack>
  );
}
