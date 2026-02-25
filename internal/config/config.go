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
	Env           string        `yaml:"env" `
	LogLevel      string        `yaml:"log_level"`
	HTTPServer    HTTPServer    `yaml:"http_server"`
	Db            Db            `yaml:"db"`
	Jwt           Jwt           `yaml:"jwt"`
	GRPC          GRPC          `yaml:"grpc"`
	MetricsServer MetricsServer `yaml:"metric_server"`
	SwaggerServer SwaggerServer `yaml:"swagger_server"`
}

type GRPC struct {
	Address     string        `yaml:"address"`
	MaxConnIdle time.Duration `yaml:"maxConnIdle"`
	MaxConnAge  time.Duration `yaml:"maxConnAge"`
}

type HTTPServer struct {
	Address               string        `yaml:"address"`
	ReadTimeout           time.Duration `yaml:"read_timeout"`
	ReadHeaderTimeout     time.Duration `yaml:"read_header_timeout"`
	WriteTimeout          time.Duration `yaml:"write_timeout"`
	IdleTimeout           time.Duration `yaml:"idle_timeout"`
	BearerToken           string        `yaml:"bearer_token"`
	MaxConcurrentRequests int           `yaml:"max_concurrent_requests"`
}

type MetricsServer struct {
	Address      string        `yaml:"address"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

type SwaggerServer struct {
	Enabled      bool          `yaml:"enabled"`
	Address      string        `yaml:"address"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

type Db struct {
	Option   string `yaml:"option"`
	Driver   string `yaml:"driver"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	NameDb   string `yaml:"name_db"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`

	MaxConns        int32         `yaml:"max_conns"`
	MinConns        int32         `yaml:"min_conns"`
	MaxConnLifetime time.Duration `yaml:"max_conn_lifetime"`
	MaxConnIdleTime time.Duration `yaml:"max_conn_idle_time"`
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
		log.Printf("Error loading .env file: %v", err)
	}

	return &Config{
		Env:      MustGetDef("ENV", "test"),
		LogLevel: MustGetDef("LOG_LEVEL", "info"),

		HTTPServer: HTTPServer{
			Address:               MustGetDef("HTTP_SERVER_ADDRESS", ":8080"),
			ReadTimeout:           MustGetDef("HTTP_SERVER_READ_TIMEOUT", 5*time.Second),
			ReadHeaderTimeout:     MustGetDef("HTTP_SERVER_READ_HEADER_TIMEOUT", 3*time.Second),
			WriteTimeout:          MustGetDef("HTTP_SERVER_WRITE_TIMEOUT", 5*time.Second),
			IdleTimeout:           MustGetDef("HTTP_SERVER_IDLE_TIMEOUT", time.Minute),
			MaxConcurrentRequests: MustGetDef("HTTP_MAX_CONCURRENT_REQUESTS", 500),
		},

		MetricsServer: MetricsServer{
			Address:      MustGetDef("METRICS_SERVER_ADDRESS", ":9091"),
			ReadTimeout:  MustGetDef("METRICS_SERVER_READ_TIMEOUT", 5*time.Second),
			WriteTimeout: MustGetDef("METRICS_SERVER_WRITE_TIMEOUT", 5*time.Second),
			IdleTimeout:  MustGetDef("METRICS_SERVER_IDLE_TIMEOUT", time.Minute),
		},

		SwaggerServer: SwaggerServer{
			// NOTE:in prod, you need to turn off
			Enabled:      MustGetDef("SWAGGER_SERVER_ENABLED", true),
			Address:      MustGetDef("SWAGGER_SERVER_ADDRESS", ":8081"),
			ReadTimeout:  MustGetDef("SWAGGER_SERVER_READ_TIMEOUT", 3*time.Second),
			WriteTimeout: MustGetDef("SWAGGER_SERVER_WRITE_TIMEOUT", 5*time.Second),
			IdleTimeout:  MustGetDef("SWAGGER_SERVER_IDLE_TIMEOUT", time.Minute),
		},

		GRPC: GRPC{
			Address:     MustGetDef("GRPC_SERVER_ADDRESS", ":3000"),
			MaxConnIdle: MustGetDef("GRPC_MAX_CONN_IDLE", 5*time.Minute),
			MaxConnAge:  MustGetDef("GRPC_MAX_CONN_AGE", 10*time.Minute),
		},

		Db: Db{
			Option:   MustGetDef("DB_OPTION", "sslmode=disable"),
			Driver:   MustGetDef("DB_DRIVER", "postgres"),
			Host:     MustGetDef("DB_HOST", "postgres"),
			Port:     MustGetDef("DB_PORT", "5432"),
			NameDb:   MustGetDef("DB_NAME", "pvz-service_db"),
			User:     MustGetDef("DB_USER", "root"),
			Password: MustGetDef("DB_PASSWORD", "root"),

			MaxConns:        MustGetDef("DB_MAX_CONNS", int32(100)),
			MinConns:        MustGetDef("DB_MIN_CONNS", int32(10)),
			MaxConnLifetime: MustGetDef("DB_MAX_CONN_LIFETIME", 10*time.Minute),
			MaxConnIdleTime: MustGetDef("DB_MAX_CONN_IDLE_TIME", 5*time.Minute),
		},

		Jwt: Jwt{
			AccessLifeTime: MustGetDef("JWT_ACCESS_LIFE_TIME", 2*time.Hour),
			Iss:            MustGetDef("JWT_ISSUER", "avito-pvz-service"),
			RSAPublicFile:  MustGetDef("JWT_RSA_PUBLIC_PEM_FILE", "secrets/public.pem"),
			RSAPrivateFile: MustGetDef("JWT_RSA_PRIVATE_PEM_FILE", "secrets/private.pem"),
		},
	}
}

func MustGet[T any](key string) T {
	raw := os.Getenv(key)
	if raw == "" {
		log.Fatalf("env %s is not set", key)
	}

	result, err := parseEnvValue[T](key, raw)
	if err != nil {
		log.Fatalf("env %s: %v", key, err)
	}
	return result
}

func MustGetDef[T any](key string, def T) T {
	raw := os.Getenv(key)
	if raw == "" {
		return def
	}
	result, err := parseEnvValue[T](key, raw)
	if err != nil {
		log.Fatalf("env %s: %v", key, err)
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

	case int32:
		v, err := strconv.ParseInt(raw, 10, 32)
		if err != nil {
			return zero, fmt.Errorf("env %s: invalid int64 value: %w", key, err)
		}
		return any(int32(v)).(T), nil

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
