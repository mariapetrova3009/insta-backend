package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
)

type ctxKey string

const ctxUserID ctxKey = "user-id"

// get Authorization
func MetadataFromHTTP(r *http.Request) metadata.MD {
	a := r.Header.Get("Authorization")
	md := metadata.MD{}
	if a != "" {
		md.Set("authorization", a)
	}
	if uid, ok := r.Context().Value("user-id").(string); ok && uid != "" {
		md.Set("user-id", uid)
	}
	return md // <— было metadata.Pairs(...)
}

// connect context with metadata
func Outgoing(ctx context.Context, md metadata.MD) context.Context {
	if md == nil || len(md) == 0 {
		return ctx
	}
	return metadata.NewOutgoingContext(ctx, md)
}

// HTTP-middleware for check token`s Bearer
func JWTMiddleware(secret []byte) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a := r.Header.Get("Authorization")
			if a == "" || !strings.HasPrefix(strings.ToLower(a), "bearer ") {
				http.Error(w, "missing token", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimSpace(a[len("Bearer "):])
			tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				// при необходимости проверь alg (например, t.Method.Alg() == jwt.SigningMethodHS256.Alg())
				return secret, nil
			})
			if err != nil || !tok.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			claims := tok.Claims.(jwt.MapClaims)
			uid, _ := claims["sub"].(string)
			ctx := context.WithValue(r.Context(), "user-id", uid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})

	}
}
