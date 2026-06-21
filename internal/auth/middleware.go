package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/protocol"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			protocol.WriteJSON(w, http.StatusUnauthorized, protocol.ErrorResponse{Error: "missing token"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return a.jwtSecret, nil
		})
		if err != nil {
			protocol.WriteJSON(w, http.StatusUnauthorized, protocol.ErrorResponse{Error: "invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			protocol.WriteJSON(w, http.StatusUnauthorized, protocol.ErrorResponse{Error: "invalid claims"})
			return
		}

		userID, ok := claims["sub"].(string)
		if !ok || userID == "" {
			protocol.WriteJSON(w, http.StatusUnauthorized, protocol.ErrorResponse{Error: "invalid token"})
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
