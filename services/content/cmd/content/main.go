package main

import (
	"context"
	"net"

	"github.com/mariapetrova3009/insta-backend/services/content/internal/server"

	"github.com/mariapetrova3009/insta-backend/services/content/internal/storage"

	"github.com/jackc/pgx/v5/pgxpool"
	cfgpkg "github.com/mariapetrova3009/insta-backend/pkg/config"
	logpkg "github.com/mariapetrova3009/insta-backend/pkg/logger"
	contentpb "github.com/mariapetrova3009/insta-backend/proto/content"
	"google.golang.org/grpc"
)

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

const service = "content"

func main() {
	// 1) загружаем конфиг
	cfg, err := cfgpkg.Load(service)
	if err != nil {
		panic(err)
	}

	log := logpkg.New(cfg.Env, service, cfg.Log.Level, cfg.Log.Format)
	log.Info("starting")

	// 2) подключение к базе
	db := must(pgxpool.New(context.Background(), cfg.Postgres.DSN))
	defer db.Close()

	st := storage.NewLocalFS(cfg.Storage.UploadDir)
	s := server.New(db, st)

	// 3) gRPC
	grpcSrv := grpc.NewServer()
	contentpb.RegisterContentServiceServer(grpcSrv, s)

	lis := must(net.Listen("tcp", cfg.GRPC.Addr))
	log.Info("grpc listen", "addr", cfg.GRPC.Addr)
	must(0, grpcSrv.Serve(lis))
}
