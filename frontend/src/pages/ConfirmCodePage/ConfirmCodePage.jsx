import { Paper, Button, Title, Text, TextInput, Stack, Group } from "@mantine/core";
import { useForm } from "@mantine/form";
import { useState } from "react";
import { useLocation, useNavigate } from "react-router";
import { confirmPasswordReset } from "../../api/auth";
import { validatePassword, validateResetCode } from "../../utils/authValidation";
import classes from "./ConfirmCodePage.module.css";

export function ConfirmCodePage(){
    const navigate = useNavigate();
    const location = useLocation();
    const searchParams = new URLSearchParams(location.search);
    const initialEmail = location.state?.email ?? searchParams.get("email") ?? "";
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [apiError, setApiError] = useState("");

    const form = useForm({
        mode: "uncontrolled",
        initialValues: {
            email: initialEmail,
            code: "",
            password: "",
            confirmPassword: "",
        },
        validate: {
            code: validateResetCode,
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
        },
    });

    async function handleSubmit(values){
        try{
            setApiError("");
            setIsSubmitting(true);

            await confirmPasswordReset({
                email: values.email.trim(),
                code: values.code.trim(),
                newPassword: values.password,
            });

            navigate("/", { replace: true });
        }catch(error){
            const fieldMap = {
                email: "email",
                code: "code",
                password: "password",
                new_password: "password",
                newPassword: "password",
            };
            const fieldName = fieldMap[error.field];

            if (fieldName){
                form.setFieldError(fieldName, error.message);
                return;
            }

            setApiError(error.message || "Не удалось сбросить пароль");
        }finally{
            setIsSubmitting(false);
        }
    }

    return(
        <Paper component="form" onSubmit={form.onSubmit(handleSubmit)} className={classes.Card}>
            <Stack gap={0} className={classes.ContentStack}>
                <Group className={classes.CardHeader}>
                    <button
                        type="button"
                        className={classes.ButtonExit}
                        onClick={() => navigate("/reset-password")}
                    >
                        <span className="material-symbols-outlined">
                            chevron_left
                        </span>
                        Назад
                    </button>
                    <Title className={classes.TitleText}>
                        Сброс пароля
                    </Title>
                </Group>
                <Text className={classes.Text}>
                    На вашу почту направлено письмо с кодом
                </Text>
                <TextInput
                    className={classes.TextInput}
                    placeholder="Введите код"
                    classNames={{
                        wrapper: classes.InputWrapper,
                        error: classes.FieldError,
                    }}
                    key={form.key("code")}
                    {...form.getInputProps("code")}
                />
                <Text className={classes.FieldLabel}>
                    Новый пароль
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
                <Text className={classes.FieldLabel}>
                    Подтвердите пароль
                </Text>
                <TextInput
                    type="password"
                    className={`${classes.TextInput} ${classes.LastTextInput}`}
                    placeholder="Введите пароль"
                    classNames={{
                        wrapper: classes.InputWrapper,
                        error: classes.FieldError,
                    }}
                    key={form.key("confirmPassword")}
                    {...form.getInputProps("confirmPassword")}
                />
                {apiError && (
                    <Text className={classes.ApiError}>
                        {apiError}
                    </Text>
                )}
                <Button type="submit" loading={isSubmitting} className={classes.Button}>
                    Проверить
                </Button>
            </Stack>
        </Paper>
    );
}
