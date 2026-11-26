package middleware

import (
	"context"
	"library-management-api/internal/response"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type ctxKey string

const (
	CtxUserID   ctxKey = "user_id"
	CtxUserRole ctxKey = "user_role"
)

func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется заголовок Authorization: Bearer <token>")
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "super-secret-jwt-key-change-in-production"
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			response.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "Недействительный или просроченный токен")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Error(w, http.StatusUnauthorized, "INVALID_CLAIMS", "Некорректные данные токена")
			return
		}

		userIDFloat, ok1 := claims["user_id"].(float64)
		role, ok2 := claims["role"].(string)
		if !ok1 || !ok2 {
			response.Error(w, http.StatusUnauthorized, "INVALID_CLAIMS", "В токене отсутствуют идентификатор или роль")
			return
		}

		ctx := context.WithValue(r.Context(), CtxUserID, int(userIDFloat))
		ctx = context.WithValue(ctx, CtxUserRole, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
