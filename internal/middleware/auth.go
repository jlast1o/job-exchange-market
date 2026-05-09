package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/jlast1o/job-exchange/internal/auth"
)

type contextKey string

const userContextKey contextKey = "user"

type UserInfo struct {
	ID    int64
	Email string
	Role  string
}

func AuthMiddleware(tokenValidator *auth.TokenGenerator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Потерян header авторизации", http.StatusUnauthorized)
				return
			}
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, "Неверный формат авторизации", http.StatusUnauthorized)
				return
			}

			claims, err := tokenValidator.ValidateToken(parts[1])
			if err != nil {
				http.Error(w, "Инвалидный или просроченный токен", http.StatusUnauthorized)
				return
			}

			user := &UserInfo{
				ID:    claims.UserID,
				Email: claims.Email,
				Role:  claims.Role,
			}
			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext извлекает пользователя из контекста.
func GetUserFromContext(ctx context.Context) (*UserInfo, bool) {
	user, ok := ctx.Value(userContextKey).(*UserInfo)
	return user, ok
}
