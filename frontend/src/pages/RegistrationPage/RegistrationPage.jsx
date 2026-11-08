import { Stack, Paper, Button, Text, Title, TextInput, Group, Checkbox } from "@mantine/core";
import { useForm } from "@mantine/form";
import { useState } from "react";
import { useNavigate } from "react-router";
import { registerUser } from "../../api/auth";
import { validateEmail, validatePassword } from "../../utils/authValidation";
import classes from "./RegistrationPage.module.css";

export function RegistrationPage(){
    const navigate = useNavigate();
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [apiError, setApiError] = useState("");

    const form = useForm({
        mode: "uncontrolled",
        initialValues: {
            email: "",
            password: "",
            confirmPassword: "",
            agreement: false,
        },
        validate: {
            email: validateEmail,
            password: validatePassword,
            confirmPassword: (value, values) => {
                if (String(value ?? "").trim() === ""){
                    return "Подтвердите пароль";
                }

                if (value !== values.password){
                    return "Пароли не совпадают";
                }

                return null;
            },
            agreement: (value) => {
                if (!value){
                    return "Подтвердите согласие";
                }

                return null;
            },
        },
    });

    async function handleSubmit(values){
        try{
            setApiError("");
            setIsSubmitting(true);

            await registerUser({
                email: values.email.trim(),
                password: values.password,
            });

            navigate("/projects", { replace: true });
        }catch(error){
            const fieldMap = {
                password: "password",
                email: "email",
            };
            const fieldName = fieldMap[error.field];

            if (fieldName){
                form.setFieldError(fieldName, error.message);
                return;
            }

            setApiError(error.message || "Не удалось зарегистрироваться");
        }finally{
            setIsSubmitting(false);
        }
    }

    return(
        <Paper component="form" onSubmit={form.onSubmit(handleSubmit)} className={classes.Card}>
            <Stack gap={0} className={classes.ContentStack}>
                <Title className={classes.TitleText}>
                    Регистрация
                </Title>
                <Group className={classes.Group} gap={4} align="baseline" justify="center">
                    <Text className={classes.Text}>
                        или
                    </Text>
                    <Button
                        type="button"
                        variant="transparent"
                        className={classes.Button}
                        onClick={() => navigate("/")}
                    >
                        войдите в свой аккаунт
                    </Button>
                    <Text className={classes.AfterText}>
                        , если он уже есть
                    </Text>
                </Group>
                <Text className={classes.UnderText}>
                    E-mail<span className={classes.Star}>*</span>
                </Text>
                <TextInput
                    type="email"
                    className={classes.TextInput}
                    placeholder="Введите ваш e-mail"
                    classNames={{
                        wrapper: classes.InputWrapper,
                        error: classes.FieldError,
                    }}
                    key={form.key("email")}
                    {...form.getInputProps("email")}
                />
                <Text className={classes.UnderText}>
                    Пароль<span className={classes.Star}>*</span>
                </Text>
                <TextInput
                    type="password"
                    className={classes.TextInput}
                    placeholder="Введите пароль"
                    classNames={{
                        wrapper: classes.InputWrapper,
                        error: classes.FieldError,
                    }}
                    key={form.key("password")}
                    {...form.getInputProps("password")}
                />
                <Text className={classes.UnderText}>
                    Подтвердите пароль<span className={classes.Star}>*</span>
                </Text>
                <TextInput
                    type="password"
                    className={classes.TextInput}
                    placeholder="Введите пароль"
                    classNames={{
                        wrapper: classes.InputWrapper,
                        error: classes.FieldError,
                    }}
                    key={form.key("confirmPassword")}
                    {...form.getInputProps("confirmPassword")}
                />
                <Checkbox
                    label="Согласен(-а) на обработку персональных данных"
                    color="dark"
                    classNames={{
                        input: classes.CheckboxInput,
                        label: classes.CheckboxLabel,
                        body: classes.CheckboxBody,
                        inner: classes.CheckboxInner,
                        error: classes.FieldError,
                    }}
                    key={form.key("agreement")}
                    {...form.getInputProps("agreement", { type: "checkbox" })}
                />
                {apiError && (
                    <Text className={classes.ApiError}>
                        {apiError}
                    </Text>
                )}
                <Button type="submit" loading={isSubmitting} className={classes.ButtonRegister}>
                    Зарегистрироваться
                </Button>
            </Stack>
        </Paper>
    );
}
