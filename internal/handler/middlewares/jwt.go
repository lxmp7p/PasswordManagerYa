package middlewares

import (
	"context"
	"net/http"
	"passwordmanager/internal/service"
	"strconv"
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

			userID, err := strconv.ParseInt(claims.Subject, 10, 64)
			if err != nil {
				http.Error(w, "invalid user id", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(
				r.Context(),
				UserIDKey,
				userID,
			)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
}
