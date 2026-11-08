package httpapi

import "lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/domain"

type AuthRegisterRequestDTO struct {
	Email       string  `json:"email"`
	Password    string  `json:"password"`
	DisplayName *string `json:"display_name"`
}

type AuthLoginRequestDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthUserDTO struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	DisplayName *string `json:"display_name"`
}

type AuthResponseDTO struct {
	User AuthUserDTO `json:"user"`
}

func authUserToDTO(user domain.User) AuthUserDTO {
	return AuthUserDTO{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
	}
}

func authResponseToDTO(user domain.User) AuthResponseDTO {
	return AuthResponseDTO{User: authUserToDTO(user)}
}

type PasswordResetRequestDTO struct {
	Email string `json:"email"`
}

type PasswordResetConfirmDTO struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}
