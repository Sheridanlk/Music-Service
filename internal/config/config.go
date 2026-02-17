package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string      `yaml:"env"`
	HTTPServer  HTTPServer  `yaml:"http_server"`
	PostgreSQL  PostgreSQL  `yaml:"postgresql"`
	MinIOClient MinIOClient `yaml:"minio_client"`
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
	Password string `yaml:"password" env_required:"PGSQL_PASSWORD"`
}

type MinIOClient struct {
	endpoint        string `yaml:"endpoint"`
	accessKeyID     string `yaml:"access_key_id"`
	secretAccessKey string `yaml:"secret_access_key" env_required:"MINIO_SECRET_ACCESS_KEY"`
	useSSL          bool   `yaml:"use_ssl"`
}

func Load() *Config {
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
