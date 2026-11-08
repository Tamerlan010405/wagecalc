import { Paper, Text, TextInput, Button, Stack, Title, Group } from "@mantine/core";
import { useForm } from "@mantine/form";
import { useState } from "react";
import { useNavigate } from "react-router";
import { requestPasswordReset } from "../../api/auth";
import { validateEmail } from "../../utils/authValidation";
import classes from  "./ResetPasswordPage.module.css";

export function ResetPasswordPage(){
    const navigate = useNavigate();
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [apiError, setApiError] = useState("");

    const form = useForm({
        mode: "uncontrolled",
        initialValues: {
            email: "",
        },
        validate: {
            email: validateEmail,
        },
    });

    async function handleSubmit(values){
        const email = values.email.trim();

        try{
            setApiError("");
            setIsSubmitting(true);

            await requestPasswordReset({ email });

            navigate(`/confirm-code?email=${encodeURIComponent(email)}`, {
                state: {
                    email,
                },
            });
        }catch(error){
            if (error.field === "email"){
                form.setFieldError("email", error.message);
                return;
            }

            setApiError(error.message || "Не удалось отправить код");
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
                        onClick={() => navigate("/")}
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
                    E-mail
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
                {apiError && (
                    <Text className={classes.ApiError}>
                        {apiError}
                    </Text>
                )}
                <Button type="submit" loading={isSubmitting} className={classes.Button}>
                    Отправить
                </Button>
            </Stack>
        </Paper>
    );
}
