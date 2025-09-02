package http

import (
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	cfgpkg "github.com/mariapetrova3009/insta-backend/pkg/config"
	"github.com/mariapetrova3009/insta-backend/services/gateway/internal/auth"
	"github.com/mariapetrova3009/insta-backend/services/gateway/internal/clients"
)

func NewRouter(log *slog.Logger, cfg *cfgpkg.Config, cl *clients.Clients) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer, middleware.Timeout(15*time.Second))

	// handlers.go в том же пакете, поэтому просто вызываем функции без префикса
	r.Get("/healthz", Healthz())

	// auth
	r.Post("/auth/register", Register(cl))
	r.Post("/auth/login", Login(cl))

	r.With(auth.JWTMiddleware([]byte(cfg.JWT.Secret))).Group(func(pr chi.Router) {
		pr.Get("/me", Me(cl))
		pr.Post("/posts", CreatePost(cl)) // multipart: file + caption
	})

	r.Get("/feed", GetFeed(cl))
	return r
}
