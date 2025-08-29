package main

import (
	"net/http"

	cfgpkg "github.com/mariapetrova3009/insta-backend/pkg/config"
	logpkg "github.com/mariapetrova3009/insta-backend/pkg/logger"

	gatewayclients "github.com/mariapetrova3009/insta-backend/services/gateway/internal/clients"
	gatewayhttp "github.com/mariapetrova3009/insta-backend/services/gateway/internal/http"
)

const service = "gateway"

func main() {
	cfg, err := cfgpkg.Load(service)
	if err != nil {
		panic(err)
	}
	log := logpkg.New(cfg.Env, service, cfg.Log.Level, cfg.Log.Format)

	cl := gatewayclients.MustInit(cfg)
	r := gatewayhttp.NewRouter(log, cfg, cl)

	log.Info("http listen", "addr", cfg.HTTP.Addr)
	if err := http.ListenAndServe(cfg.HTTP.Addr, r); err != nil {
		log.Error("server error", "err", err)
	}
}
