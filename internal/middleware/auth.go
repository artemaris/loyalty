package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/artemaris/loyalty/internal/services"
)

type AuthMiddleware struct {
	authService *services.AuthService
}

// Кастомные типы для ключей контекста
type contextKey string

const (
	userIDKey    contextKey = "user_id"
	userLoginKey contextKey = "user_login"
)

func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// Authenticate middleware для проверки JWT токена
func (a *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string

		// 1. Пробуем из заголовка Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		// 2. Если не нашли — пробуем из cookie
		if tokenString == "" {
			cookie, err := r.Cookie("auth_token")
			if err == nil {
				tokenString = cookie.Value
			}
		}

		if tokenString == "" {
			http.Error(w, "Authorization required", http.StatusUnauthorized)
			return
		}

		claims, err := a.authService.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		ctx = context.WithValue(ctx, userLoginKey, claims.Login)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext получает ID пользователя из контекста
func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDKey).(int64)
	return userID, ok
}

// GetUserLoginFromContext получает логин пользователя из контекста
func GetUserLoginFromContext(ctx context.Context) (string, bool) {
	login, ok := ctx.Value(userLoginKey).(string)
	return login, ok
}
