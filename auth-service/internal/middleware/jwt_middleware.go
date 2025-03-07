package middleware

import (
	"auth-service/internal/controller"
	"auth-service/internal/service"
	"context"
	"net/http"
	"strings"
)

type JWTMiddleware struct {
	JWTService *service.JWTService
}

func NewJWTMiddleware(jwtService *service.JWTService) *JWTMiddleware {
	return &JWTMiddleware{JWTService: jwtService}
}

func (m *JWTMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			controller.SendErrorResponse(w, http.StatusUnauthorized, "Missing Authorization header")
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			controller.SendErrorResponse(w, http.StatusUnauthorized, "Invalid Authorization header format")
			return
		}

		userID, err := m.JWTService.ValidateAccessToken(parts[1])
		if err != nil {
			controller.SendErrorResponse(w, http.StatusUnauthorized, "Invalid token")
			return
		}
		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
