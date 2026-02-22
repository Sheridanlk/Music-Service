package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env          string       `yaml:"env"`
	HTTPServer   HTTPServer   `yaml:"http_server"`
	PostgreSQL   PostgreSQL   `yaml:"postgresql"`
	MinIOClient  MinIOClient  `yaml:"minio_client"`
	MinioStorage MinioStorage `yaml:"minio_storage"`
}

type HTTPServer struct {
	Host        string        `yaml:"host"`
	Port        int           `yaml:"port"`
	Timeout     time.Duration `yaml:"timeout"`
	IdleTimeout time.Duration `yaml:"idle_timeout"`
}

type PostgreSQL struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	DBName   string `yaml:"db_name"`
	UserName string `yaml:"user_name"`
	Password string `yaml:"password" env:"PGSQL_PASSWORD" env_required:"true"`
}

type MinIOClient struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key" env:"MINIO_SECRET_ACCESS_KEY" env_required:"true"`
	UseSSL          bool   `yaml:"use_ssl"`
}

type MinioStorage struct {
	OriginalBucket string `yaml:"original_bucket"`
	HLSBucket      string `yaml:"hls_bucket"`
}

func Load() *Config {
	_ = godotenv.Load()

	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		log.Fatal("CONFIG_PATH environment variable is not set")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatal("Config file does not exist at path:", path)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		log.Fatal("Failed to read config file:", err)
	}

	return &cfg
}
