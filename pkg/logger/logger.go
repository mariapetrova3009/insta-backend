package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New создаёт настроенный slog-логгер.
// env: "local/dev/prod" — включим AddSource вне prod.
// service: имя сервиса (identity/gateway/...), попадёт в поля.
// level: "debug|info|warn|error"
// format: "console|json"
func New(env, service, level, format string) *slog.Logger {
	lvl := parseLevel(level)
	opts := &slog.HandlerOptions{
		Level:     lvl,
		AddSource: strings.ToLower(env) != "prod",
	}

	var h slog.Handler
	if strings.ToLower(format) == "json" {
		h = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		h = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(h).With(
		"service", service,
		"env", env,
	)

	slog.SetDefault(logger)

	return logger
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
