package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	cfgpkg "github.com/mariapetrova3009/insta-backend/pkg/config"
	logpkg "github.com/mariapetrova3009/insta-backend/pkg/logger"
	contentpb "github.com/mariapetrova3009/insta-backend/proto/content"
	contentrepo "github.com/mariapetrova3009/insta-backend/services/content/internal/repo"
	contentserver "github.com/mariapetrova3009/insta-backend/services/content/internal/server"
	contentstore "github.com/mariapetrova3009/insta-backend/services/content/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const service = "content"

func main() {
	// config
	cfg, err := cfgpkg.Load(service)
	if err != nil {
		panic(err)
	}

	// logger
	log := logpkg.New(cfg.Env, service, cfg.Log.Level, cfg.Log.Format)
	log.Info("starting")

	// connect to db
	db, err := sql.Open("postgres", cfg.Postgres.DSN)
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	// HTTP

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
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

	repo := contentrepo.NewRepo(db)
	store := contentstore.NewLocalFS(cfg.Storage.UploadDir)
	prod, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":  strings.Join(cfg.Kafka.Brokers, ","),
		"enable.idempotence": true,
		"acks":               "all",
		"linger.ms":          10,
		"retries":            5,
	})
	if err != nil {
		log.Error("kafka init", "err", err)
		return
	}
	defer prod.Close()

	srv := contentserver.New(log, repo, store, prod, cfg.Kafka.Topics.PostCreated)
	contentpb.RegisterContentServiceServer(grpcSrv, srv)

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
	case err := <-errCh:
		log.Error("server error", slog.Any("err", err))
	case sig := <-stop:
		log.Info("stopping", "signal", sig.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	httpSrv.Shutdown(ctx)
	grpcSrv.GracefulStop()
	log.Info("stopped")
}
