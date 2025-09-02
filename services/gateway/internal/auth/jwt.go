package auth

import (
	"context"
	"net/http"
	"strings"

	"google.golang.org/grpc/metadata"
)

// get Authorization
func MetadataFromHTTP(r *http.Request) metadata.MD {
	a := r.Header.Get("Authorization")
	if a == "" {
		return metadata.MD{}
	}

	return metadata.Pairs("authorization", a)
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
			next.ServeHTTP(w, r)
		})
	}
}
