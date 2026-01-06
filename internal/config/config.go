package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env        string `yaml:"env" `
	LogLevel   string `yaml:"log_level"`
	HTTPServer `yaml:"http_server"`
	Db         `yaml:"db"`
}

type HTTPServer struct {
	Address     string `yaml:"address"`
	Timeout     int    `yaml:"timeout"`
	BearerToken string `yaml:"bearer_token"`
	DbPort      string `yaml:"db_port"`
}

type Db struct {
	Option       string `yaml:"option"`
	Driver       string `yaml:"driver"`
	Host         string `yaml:"host"`
	ExternalPort string `yaml:"port"`
	InternalPort string `yaml:"port"`
	NameDb       string `yaml:"name_db"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
}

func LoadConfig(configPath string) *Config {
	var err error

	if configPath == "" {
		err = godotenv.Load()
	} else {
		err = godotenv.Load(configPath)
	}

	if err != nil {
		log.Fatal("Error loading .env file: %w", err)
	}

	return &Config{
		Env:      MustGet("ENV", "test"),
		LogLevel: MustGet("LOG_LEVEL", "info"),
		HTTPServer: HTTPServer{
			Address: MustGet("HTTP_SERVER_HOST", "") + ":" + MustGet("HTTP_SERVER_PORT", ""),
			Timeout: MustGet("HTTP_TIMEOUT", 0),
			DbPort:  MustGet("HTTP_DB_PORT", ""),
		},
		Db: Db{
			Option:       MustGet("DB_OPTION", ""),
			Driver:       MustGet("DB_DRIVER", ""),
			Host:         MustGet("DB_HOST", ""),
			ExternalPort: MustGet("DB_EXTERNAL_PORT", ""),
			InternalPort: MustGet("DB_INTERNAL_PORT", ""),
			NameDb:       MustGet("DB_NAME", ""),
			User:         MustGet("DB_USER", ""),
			Password:     MustGet("DB_PASSWORD", ""),
		},
	}
}

func MustGetDef[T any](key string, def T) T {
	raw := os.Getenv(key)
	if raw == "" {
		return def
	}

	return MustGet(key, def)
}

func MustGet[T any](key string, def T) T {
	raw := os.Getenv(key)
	if raw == "" {
		return def
	}

	var result T

	switch any(result).(type) {
	case string:
		return any(raw).(T)

	case int:
		v, err := strconv.Atoi(raw)
		if err != nil {
			log.Fatalf("env %s: invalid int value: %v", key, err)
		}
		return any(v).(T)

	case int64:
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			log.Fatalf("env %s: invalid int64 value: %v", key, err)
		}
		return any(v).(T)

	case bool:
		v, err := strconv.ParseBool(raw)
		if err != nil {
			log.Fatalf("env %s: invalid bool value: %v", key, err)
		}
		return any(v).(T)

	case time.Duration:
		v, err := time.ParseDuration(raw)
		if err != nil {
			log.Fatalf("env %s: invalid duration value: %v", key, err)
		}
		return any(v).(T)

	default:
		log.Fatalf("env %s: unsupported type %T", key, result)
	}

	panic("unreachable")
}
