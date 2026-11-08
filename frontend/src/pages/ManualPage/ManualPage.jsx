import { Text, Title, Stack, Paper } from "@mantine/core";
import classes from "./ManualPage.module.css"

import { Link, useLocation } from "react-router";


export function ManualPage(){
    const location = useLocation();
    const routeState = location.state ?? {};
    const reportTitle = routeState.reportTitle;
    const reportPath = routeState.reportPath ?? "/projects";

    return(
        <>
            <div className={classes.Breadcrumbs}>
                <Text
                    component={Link}
                    to="/projects"
                    size="sm"
                    className={classes.BreadcrumbLink}
                    >
                        Главная
                </Text>
                {reportTitle && (
                    <>
                        <Text size="sm" c="dimmed">
                            ›
                        </Text>
                        <Text
                            component={Link}
                            to={reportPath}
                            size="sm"
                            className={classes.BreadcrumbLink}
                            >
                                {reportTitle}
                        </Text>
                    </>
                )}
            <Text size="sm" c="dimmed">
            ›
            </Text>
                <Text size="sm">
                    Методика и допущения
                </Text>
            </div>
        <Stack gap={12}>
            <Title className={classes.TitleText}>
                    Методика и допущения
            </Title>
            <Paper className={classes.Card}>
                <Text className={classes.UnderTitleText}>
                    Что сравнивается
                </Text>
                <Text className={classes.Text1}>
                    Среднемесячный размер выплат по организации за последние 3 месяца последнего отчётного (расчётного) периода должен быть не ниже среднемесячной
                    <br /> 
                    начисленной ЗП по РФ или по субъекту регистрации — за квартал, предшествующий этому периоду (пп. «б» п. 4 Положения Nº 1729 в ред. 2025 г.).
                    <br /> 
                    Источник планки — ЕМИСС Росстата; выбирается более выгодная из двух величин.
                </Text>
                <Text className={classes.UnderTitleText}>
                    Как считается размер выплат (п. 5 + письмо ФНС)
                </Text>
                <Text className={classes.Text2}>
                    Из Раздела 3 каждого РСВ за квартал по каждому застрахованному лицу берётся 140 - 160 — выплаты по трудовым договорам без ГПХ. Далее:
                    <br />
                    1. За каждый месяц: сумма (140-160) | по всем физлицам ÷ число физлиц, у которых (140-160) {">"} 0 — средняя за месяц.
                    <br />
                    2. Среднемесячная за квартал = сумма месячных средних ÷ число месяцев, в которых были выплаты (от 1 до 3).
                    <br />
                    Если 140-160 = 0 у всех за все три месяца — показатель равен 0 ₽.
                </Text>
                <Text className={classes.UnderTitleText}>
                    Допущения этой реализации   
                </Text>
                <Text className={classes.Text3}>
                    1. Одна строка = одно физлицо (один СНИЛС) на весь квартал; анонимно. • Если ГПХ {">"} всего, выплата по ТД считается нулевой (140 ≥ 160 по природе формы).
                    <br />
                    2. Декретные/нулевые месяцы у сотрудника не попадают в знаменатель численности этого месяца, т.к. учитываются только выплаты {">"} О. 
                    <br />
                    3. Планка вводится вручную как среднемесячная за нужный квартал (по «запоздалости» данных Росстата — здесь это снимается ручным вводом и режимом «что нужно добить»).
                    <br />
                    <br />
                    Источники: Постановление Правительства РФ Nº 1729 (ред. 01.06.2025), пп. «б» п. 4 и п. 5; письмо ФНС от 28.11.2022 Nº 3Г-3-11/13312; данные ЕМИСС Росстата (среднемесячная начисленная заработная плата).
                </Text>
            </Paper>
        </Stack>
        </>
    );
}
