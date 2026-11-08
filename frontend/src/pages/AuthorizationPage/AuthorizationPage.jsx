import { Paper, Stack, Title, Text, Button, Group, TextInput, Divider } from "@mantine/core";
import { useForm } from "@mantine/form";
import { useState } from "react";
import { useNavigate } from "react-router";
import { loginUser } from "../../api/auth";
import { validateEmail } from "../../utils/authValidation";
import classes from "./AuthorizationPage.module.css";

export function AuthorizationPage(){
    const navigate = useNavigate();
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [apiError, setApiError] = useState("");

    const form = useForm({
        mode: "uncontrolled",
        initialValues: {
            email: "",
            password: "",
        },
        validate: {
            email: validateEmail,
            password: (value) => {
                if (String(value ?? "").trim() === ""){
                    return "Введите пароль";
                }

                return null;
            },
        },
    });

    async function handleSubmit(values){
        try{
            setApiError("");
            setIsSubmitting(true);

            await loginUser({
                email: values.email.trim(),
                password: values.password,
            });

            navigate("/projects", { replace: true });
        }catch(error){
            if (error.field){
                form.setFieldError(error.field, error.message);
                return;
            }

            setApiError(error.message || "Не удалось войти в аккаунт");
        }finally{
            setIsSubmitting(false);
        }
    }

    return(
        <Paper component="form" onSubmit={form.onSubmit(handleSubmit)} className={classes.Card}>
            <Stack className={classes.ContentStack}>
                <Title className={classes.TitleText}>
                    Вход в аккаунт
                </Title>
                <Group className={classes.GroupText} gap={3} align="center">
                    <Text className={classes.Text}>
                        Нет аккаунта?
                    </Text>
                    <Button
                        type="button"
                        className={classes.Button}
                        variant="transparent"
                        onClick={() => navigate("/registration")}
                    >
                        Зарегистрируйтесь
                    </Button>
                </Group>
                <Stack gap={4} className={classes.FieldGroup}>
                    <Text className={classes.UnderText}>
                        E-mail
                    </Text>
                    <TextInput
                        type="email"
                        placeholder="Введите ваш e-mail"
                        className={classes.TextInput}
                        classNames={{
                            wrapper: classes.InputWrapper,
                            error: classes.FieldError,
                        }}
                        key={form.key("email")}
                        {...form.getInputProps("email")}
                    />
                </Stack>
                <div className={classes.PasswordBlock}>
                    <Stack gap={4}>
                        <Text className={classes.UnderText}>
                            Пароль
                        </Text>
                        <TextInput
                            type="password"
                            placeholder="Введите пароль"
                            className={classes.TextInput}
                            classNames={{
                                wrapper: classes.InputWrapper,
                                error: classes.FieldError,
                            }}
                            key={form.key("password")}
                            {...form.getInputProps("password")}
                        />
                    </Stack>
                    <Button
                        type="button"
                        className={classes.ButtonPassword}
                        variant="transparent"
                        onClick={() => navigate("/reset-password")}
                    >
                        Забыли пароль?
                    </Button>
                </div>
                <Divider className={classes.Divider} />
                {apiError && (
                    <Text className={classes.ApiError}>
                        {apiError}
                    </Text>
                )}
                <Button type="submit" loading={isSubmitting} className={classes.ButtonEntrance}>
                    Войти
                </Button>
            </Stack>
        </Paper>
    );
}
