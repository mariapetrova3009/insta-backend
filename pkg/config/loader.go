package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func Load(service string) (*Config, error) {
	env := getenv("APP_ENV", "local")         // local|dev|prod …
	cfgDir := getenv("CONFIG_DIR", "configs") // где лежат yaml'ы

	v := viper.New()
	v.SetConfigType("yaml")

	// ENV перетирают файлы: IDENTITY_GRPC_ADDR, GATEWAY_HTTP_ADDR, CONTENT_STORAGE_UPLOAD_DIR и т.д.
	v.SetEnvPrefix(strings.ToUpper(service))
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 1) base.yaml (общие настройки)
	basePath := filepath.Join(cfgDir, "base.yaml")
	v.SetConfigFile(basePath)
	_ = v.ReadInConfig() // норм, если файла нет

	// 2) <service>/<env>.yaml (частные настройки)
	svcPath := filepath.Join(cfgDir, strings.ToLower(service), env+".yaml")
	v.SetConfigFile(svcPath)
	_ = v.MergeInConfig() // тоже можно без файла

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}
	c.Env = env
	return &c, nil
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
