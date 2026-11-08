import { Paper, Text, Button, Stack, Group } from "@mantine/core";
import { useNavigate } from "react-router";

import classes from "./ReportCard.module.css";

const QUARTER_LABELS = {
    1: "I",
    2: "II",
    3: "III",
    4: "IV",
};

function formatMoney(value){
    if (value === null || value === undefined || value === ""){
        return "–";
    }
    const number = Number(value);

    if (Number.isNaN(number)){
        return value;
    }
    return new Intl.NumberFormat("ru-RU", {
        maximumFractionDigits: 0,
    }).format(number);
}

function formatDate(value){
    if (!value){
        return "–";
    }

    const date = new Date(value);

    if (Number.isNaN(date.getTime())){
        return "–";
    }

    return new Intl.DateTimeFormat("ru-RU", {
        day: "2-digit",
        month: "2-digit",
        year: "numeric",
    }).format(date);
}

function getVerdictText(verdict){
    switch(verdict){
        case "PASS":
            return "Критерий выполняется";
        case "FAIL":
            return "Критерий не выполняется";
        case "NO_THRESHOLD":
            return "Планка не задана";
        default:
            return "Результат неизвестен";
    }
}

export function ReportCard({ report, onDelete, isDeleting = false }){

    const navigate = useNavigate();
    const quarterLabel = QUARTER_LABELS[report?.period?.quarter] ?? "–";
    const createdAt = report?.created_at;
    return(
        <Paper radius="md" className={classes.Card} h={272}>
            <Stack justify="space-between" h="100%">
                <div>
                    <Group justify="space-between">
                        <Text className={classes.HeaderText}>
                            {report?.title ?? "Без названия"}
                        </Text>
                        <Button
                            className={classes.Delete}
                            onClick={() => onDelete?.(report?.id)}
                            disabled={!report?.id || isDeleting}
                            loading={isDeleting}
                        >
                            Удалить
                        </Button>
                    </Group>
                    <Text className={classes.UnderHeaderText}>
                        Организация: <span className={classes.TextCompany}>{report?.org_name ?? "–"}</span>
                    </Text>
                    <div className={classes.InfoRow}>
                    <span className={classes.UnderHeaderText}>
                        Квартал: 
                    </span>
                    <span className={classes.Text}>
                        {quarterLabel}
                    </span>
                    </div>
                    <Text className={`${classes.TextResult} ${
                        report?.verdict === "PASS"?classes.PassResult:classes.FailResult
                    }`}>
                        {formatMoney(report?.quarter_average)} ₽/мес
                    </Text>
                    <Text className={classes.ResultTag}>
                        {getVerdictText(report?.verdict)}
                    </Text>
                </div>
                <Group justify="space-between">
                    <Text className={classes.CalendarText}>
                        <span className="material-symbols-outlined">
                            calendar_today
                        </span>
                        {formatDate(createdAt)}
                    </Text>
                    <Button 
                    onClick={() => navigate(`/calculation/${report?.id}`)}
                    disabled={!report?.id}
                    className={classes.Button}
                    classNames={{
                        inner: classes.ButtonInner
                    }}
                     variant="outline" 
                     radius="xl"
                     rightSection={
                    <span className="material-symbols-outlined">
                    arrow_forward
                    </span>
                    }
                    >
                        Смотреть отчёт
                    </Button>
                </Group>
            </Stack>
        </Paper>
    );
}
