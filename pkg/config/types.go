package config

import "time"

type Config struct {
	Env string `mapstructure:"-"`

	Log struct {
		Level  string `mapstructure:"level"`
		Format string `mapstructure:"format"`
	} `mapstructure:"log"`

	GRPC struct {
		Addr string `mapstructure:"addr"`
	} `mapstructure:"grpc"`

	HTTP struct {
		Addr string `mapstructure:"addr"`
	} `mapstructure:"http"`

	Postgres struct {
		DSN        string `mapstructure:"dsn"`
		Migrations string `mapstructure:"migrations"`
	} `mapstructure:"postgres"`

	Redis struct {
		Addr     string `mapstructure:"addr"`
		DB       int    `mapstructure:"db"`
		Password string `mapstructure:"password"`
	} `mapstructure:"redis"`

	Kafka struct {
		Brokers []string `mapstructure:"brokers"`
		Group   string   `mapstructure:"group"`
		Topics  struct {
			PostCreated string `mapstructure:"post_created"`
		} `mapstructure:"topics"`
	} `mapstructure:"kafka"`

	JWT struct {
		Secret     string        `mapstructure:"secret"`
		TTL        time.Duration `mapstructure:"ttl"`
		RefreshTTL time.Duration `mapstructure:"refresh_ttl"`
	} `mapstructure:"jwt"`

	// Хранилище медиа
	Storage struct {
		UploadDir  string        `mapstructure:"upload_dir"`
		S3         bool          `mapstructure:"s3"`
		Bucket     string        `mapstructure:"bucket"`
		Endpoint   string        `mapstructure:"endpoint"`
		AccessKey  string        `mapstructure:"access_key"`
		SecretKey  string        `mapstructure:"secret_key"`
		PresignTTL time.Duration `mapstructure:"presign_ttl"`
	} `mapstructure:"storage"`

	// Эндпоинты других сервисов
	Identity struct {
		Endpoint string `mapstructure:"endpoint"`
	} `mapstructure:"identity"`
	Content struct {
		Endpoint string `mapstructure:"endpoint"`
	} `mapstructure:"content"`
	Feed struct {
		Endpoint string `mapstructure:"endpoint"`
	} `mapstructure:"feed"`
}
