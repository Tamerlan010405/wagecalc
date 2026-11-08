import { TextInput, Select, Button, Stack } from "@mantine/core";
import { useForm } from "@mantine/form";
import classes from "./CreateReportCard.module.css";
import { useNavigate } from "react-router";

const API_BASE_URL = import.meta.env.VITE_API_URL ?? "";

function getInitialFormValues(initialValues){
  return {
    title: initialValues?.title ?? "",
    org_name: initialValues?.org_name ?? "",
    year: String(initialValues?.period?.year ?? initialValues?.year ?? ""),
    quarter: String(initialValues?.period?.quarter ?? initialValues?.quarter ?? ""),
  };
}

function getReportData(values){
  return {
    title: values.title,
    org_name: values.org_name,
    period: {
      year: Number(values.year),
      quarter: Number(values.quarter),
    },
  };
}

export function CreateReportCard({
  initialValues,
  title = "Создание отчёта",
  submitLabel = "Создать новый отчет",
  onSubmit,
}) {

  const navigate = useNavigate();

  const form = useForm({
    mode:"uncontrolled",
    initialValues: getInitialFormValues(initialValues),

  validate: {
    title: (value) =>{
      if (value.trim() === ""){
        return "Введите название отчёта";
      }else{
        return null;
      }
    },

    org_name: (value) =>{
      if (value.trim() === ""){
        return "Введите название организации"
      }else{
        return null;
      }
    },

    quarter: (value) =>{
      if (value === ""){
        return "Выберите квартал";
      }else{
        return null;
      }
    },

    year: (value) =>{
      if (value === ""){
        return "Выберите год";
      }else{
        return null;
      }
    },
  }
});

  function handleSubmit(values){
    const reportData = getReportData(values);

    if (onSubmit){
      onSubmit(reportData);
      return;
    }

    navigate("/calculation", {
      state: reportData,
    });
  }

  return (
      <form  className={classes.Form} onSubmit={form.onSubmit(handleSubmit)}>
      <Stack gap={12}>
        <div>
          <label className={classes.ReportNameLabel} htmlFor="report-name">
            {title}
          </label>
          <Stack gap={12}>
          <TextInput
            id="report-name"
            placeholder="Введите название отчёта"
            className={`${classes.Field} ${classes.NameField}`}
            classNames={{
              wrapper: classes.InputWrapper, 
              error: classes.FieldError
            }}
            key={form.key("title")}
            {...form.getInputProps("title")}
          />
          <TextInput
            id="report-organization"
            placeholder="Введите название организации"
            className={`${classes.Field} ${classes.NameField}`}
            classNames={{
              wrapper: classes.InputWrapper, 
              error: classes.FieldError
            }}
            key={form.key("org_name")}
            {...form.getInputProps("org_name")}
          />
          </Stack>
        </div>
        <Select
          placeholder="Выберите квартал"
          data={[{ value: "1", label: "1 квартал" },
                  { value: "2", label: "2 квартал" },
                  { value: "3", label: "3 квартал" },
                  { value: "4", label: "4 квартал" },]}
          classNames={{
            wrapper: classes.InputWrapper,
            error: classes.FieldError
          }}
          className={`${classes.Field} ${classes.QuarterField}`}
          key={form.key("quarter")}
          {...form.getInputProps("quarter")}
        />
        <Select
          placeholder="Выберите год"
          data={[
            {value: "2021", label: "2021"},
            {value: "2022", label: "2022"},
            {value: "2023", label: "2023"},
            {value: "2024", label: "2024"},
            {value: "2025", label: "2025"},
          ]}
          classNames={{
            wrapper: classes.InputWrapper,
            error: classes.FieldError
          }}
          className={`${classes.Field} ${classes.YearField}`}
          key={form.key("year")}
          {...form.getInputProps("year")}
        />
        <Button type="submit" radius="xl" className={classes.Button}>
        {submitLabel}
        </Button>
      </Stack>
      </form>
  );
}
