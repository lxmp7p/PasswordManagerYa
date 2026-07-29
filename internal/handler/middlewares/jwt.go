package middlewares

import (
	"context"
	"net/http"
	"passwordmanager/internal/service"
	"strings"
)

var UserIDKey = "userID"

func JWT(manager service.TokenManagerInterface) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			header := r.Header.Get("Authorization")
			if header == "" {
				http.Error(w, "missing authorization header", http.StatusBadRequest)
				return
			}

			if !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, "invalid authorization header", http.StatusBadRequest)
				return
			}

			token := strings.TrimPrefix(header, "Bearer ")
			claims, err := manager.Parse(token)

			if err != nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(
				r.Context(),
				UserIDKey,
				claims.Subject,
			)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
}
