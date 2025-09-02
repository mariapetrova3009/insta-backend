package auth

import (
	"context"
	"net/http"
	"strings"

	"google.golang.org/grpc/metadata"
)

// Достаёт Authorization из HTTP и превращает в gRPC metadata.
func MetadataFromHTTP(r *http.Request) metadata.MD {
	a := r.Header.Get("Authorization")
	if a == "" {
		return metadata.MD{}
	}
	// Можно нормализовать пробелы/регистр, но само значение не меняем
	return metadata.Pairs("authorization", a)
}

// Оборачивает context исходящими метаданными (для gRPC-клиента).
func Outgoing(ctx context.Context, md metadata.MD) context.Context {
	if md == nil || len(md) == 0 {
		return ctx
	}
	return metadata.NewOutgoingContext(ctx, md)
}

// (опционально) HTTP-middleware для проверки Bearer токена локально,
// чтобы /posts и т.п. не вызывались без токена.
func JWTMiddleware(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a := r.Header.Get("Authorization")
			if a == "" || !strings.HasPrefix(strings.ToLower(a), "bearer ") {
				http.Error(w, "missing token", http.StatusUnauthorized)
				return
			}
			// здесь можно просто пропускать дальше или валидировать подпись,
			// если хочешь – добавь парсинг через github.com/golang-jwt/jwt/v5.
			next.ServeHTTP(w, r)
		})
	}
}
