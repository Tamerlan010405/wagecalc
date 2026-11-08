package httpapi

import (
	"context"

	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/domain"
)

type authUserContextKey struct{}

func contextWithAuthUser(ctx context.Context, user domain.User) context.Context {
	return context.WithValue(ctx, authUserContextKey{}, user)
}

func AuthUserFromContext(ctx context.Context) (domain.User, bool) {
	user, ok := ctx.Value(authUserContextKey{}).(domain.User)
	return user, ok
}
