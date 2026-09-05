package auth

import (
	"net/http"
	http2 "prodApp/backend/internal/platform/http"
	"strings"
)

func AuthMiddleware(issuer *TokenIssuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				RespondUnauthorized(w)
				return
			}
			rawToken := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := issuer.Verify(rawToken)
			if err != nil {
				RespondUnauthorized(w)
				return
			}

			ctx := http2.ContextWithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := http2.ClaimsFromContext(r.Context())
		if !ok || !claims.IsAdmin {
			RespondForbidden(w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RespondUnauthorized(w http.ResponseWriter) {
	http.Error(w, "unauthorized", http.StatusUnauthorized)
}

func RespondForbidden(w http.ResponseWriter) {
	http.Error(w, "forbidden", http.StatusForbidden)
}
