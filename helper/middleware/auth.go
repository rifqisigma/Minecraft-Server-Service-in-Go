package middleware

import (
	"context"
	"minecrat_go/helper/utils"
	"net/http"
	"strings"
)

type key int

var AuthKey key = 0

func AuthMiddeware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := utils.ParseJWT(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized: Invalid or unverified user", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), AuthKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
