package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cfgpkg "github.com/mariapetrova3009/insta-backend/pkg/config"
	logpkg "github.com/mariapetrova3009/insta-backend/pkg/logger"
	fdpb "github.com/mariapetrova3009/insta-backend/proto/feed"
	feedsvc "github.com/mariapetrova3009/insta-backend/services/feed/internal/feed"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const service = "feed"

func main() {
	// config
	cfg, err := cfgpkg.Load(service)
	if err != nil {
		panic(err)
	}

	// logger
	log := logpkg.New(cfg.Env, service, cfg.Log.Level, cfg.Log.Format)
	log.Info("starting")

	db, err := sql.Open("postgres", cfg.Postgres.DSN)
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	// HTTP /healthz
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	httpSrv := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// gRPC
	grpcSrv := grpc.NewServer()
	if cfg.Env != "prod" {
		reflection.Register(grpcSrv)
	}

	repo := feedsvc.NewRepo(db)
	srv := feedsvc.New(log, repo)
	fdpb.RegisterFeedServiceServer(grpcSrv, srv)

	// run services
	errCh := make(chan error, 2)

	go func() {
		log.Info("http listen", "addr", cfg.HTTP.Addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	go func() {
		lis, err := net.Listen("tcp", cfg.GRPC.Addr)
		if err != nil {
			errCh <- err
			return
		}
		log.Info("grpc listen", "addr", cfg.GRPC.Addr)
		errCh <- grpcSrv.Serve(lis)
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		log.Info("stopping", "signal", sig.String())
	case err := <-errCh:
		log.Error("server error", slog.Any("err", err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(ctx)
	grpcSrv.GracefulStop()
	log.Info("stopped")
}
