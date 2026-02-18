package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env        string     `yaml:"env" `
	LogLevel   string     `yaml:"log_level"`
	HTTPServer HTTPServer `yaml:"http_server"`
	Db         Db         `yaml:"db"`
	Jwt        Jwt        `yaml:"jwt"`
	GRPC       GRPC       `yaml:"GRPC"`
}

type GRPC struct {
	Address     string        `yaml:"address"`
	MaxConnIdle time.Duration `yaml:"maxConnIdle"`
	MaxConnAge  time.Duration `yaml:"maxConnAge"`
}

type HTTPServer struct {
	Address      string        `yaml:"address"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
	BearerToken  string        `yaml:"bearer_token"`
	DbPort       string        `yaml:"db_port"`
}

type Db struct {
	Option   string `yaml:"option"`
	Driver   string `yaml:"driver"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	NameDb   string `yaml:"name_db"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type Jwt struct {
	AccessLifeTime time.Duration `yaml:"accessLifeTime"`
	Iss            string        `yaml:"iss"`
	RSAPublicFile  string        `yaml:"RSAPublicFile"`
	RSAPrivateFile string        `yaml:"RSAPrivateFile"`
}

func LoadConfig(configPath string) *Config {
	var err error

	if configPath == "" {
		err = godotenv.Load()
	} else {
		err = godotenv.Load(configPath)
	}

	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	return &Config{
		Env:      MustGetDef("ENV", "test"),
		LogLevel: MustGetDef("LOG_LEVEL", "info"),
		HTTPServer: HTTPServer{
			Address:      MustGet[string]("HTTP_SERVER_ADDRESS"),
			ReadTimeout:  MustGet[time.Duration]("HTTP_READ_TIMEOUT"),
			WriteTimeout: MustGet[time.Duration]("HTTP_WRITE_TIMEOUT"),
			IdleTimeout:  MustGet[time.Duration]("HTTP_IDLE_TIMEOUT"),
		},
		GRPC: GRPC{
			Address:     MustGet[string]("GRPC_SERVER_ADDRESS"),
			MaxConnIdle: MustGet[time.Duration]("GRPC_MAX_CONN_IDLE"),
			MaxConnAge:  MustGet[time.Duration]("GRPC_MAX_CONN_AGE"),
		},
		Db: Db{
			Option:   MustGet[string]("DB_OPTION"),
			Driver:   MustGet[string]("DB_DRIVER"),
			Host:     MustGet[string]("DB_HOST"),
			Port:     MustGet[string]("DB_PORT"),
			NameDb:   MustGet[string]("DB_NAME"),
			User:     MustGet[string]("DB_USER"),
			Password: MustGet[string]("DB_PASSWORD"),
		},
		Jwt: Jwt{
			AccessLifeTime: MustGet[time.Duration]("JWT_ASSERT_LIFE_TIME"),
			Iss:            MustGet[string]("JWT_ISSUER"),
			RSAPublicFile:  MustGet[string]("JWT_RSA_PUBLIC_PEM_FILE"),
			RSAPrivateFile: MustGet[string]("JWT_RSA_PRIVATE_PEM_FILE"),
		},
	}
}

func MustGet[T any](key string) T {
	raw := os.Getenv(key)
	if raw == "" {
		log.Fatalf("env %s is not set", key)
	}

	result, err := parseEnvValue[T](raw, raw)
	if err != nil {
		panic(err)
	}
	return result
}

func MustGetDef[T any](key string, def T) T {
	raw := os.Getenv(key)
	if raw == "" {
		return def
	}
	result, err := parseEnvValue[T](raw, raw)
	if err != nil {
		panic(err)
	}
	return result
}

func parseEnvValue[T any](key, raw string) (T, error) {
	var zero T

	switch any(zero).(type) {
	case string:
		return any(raw).(T), nil

	case int:
		v, err := strconv.Atoi(raw)
		if err != nil {
			return zero, fmt.Errorf("env %s: invalid int value: %w", key, err)
		}
		return any(v).(T), nil

	case int64:
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return zero, fmt.Errorf("env %s: invalid int64 value: %w", key, err)
		}
		return any(v).(T), nil

	case bool:
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return zero, fmt.Errorf("env %s: invalid bool value: %w", key, err)
		}
		return any(v).(T), nil

	case time.Duration:
		v, err := time.ParseDuration(raw)
		if err != nil {
			return zero, fmt.Errorf("env %s: invalid duration value: %w", key, err)
		}
		return any(v).(T), nil

	default:
		return zero, fmt.Errorf("env %s: unsupported type %T", key, zero)
	}
}
