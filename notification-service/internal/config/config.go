package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	DBURL      string
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBPort     string
	GRPCPort   string

	NATSURL      string
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
	DefaultEmail string
}

func Load() *Config {
	_ = godotenv.Load()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	cfg := &Config{
		DBURL:      getEnv("DB_URL", ""),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "notifications"),
		DBPort:     getEnv("DB_PORT", "5432"),
		GRPCPort:   getEnv("GRPC_PORT", "50054"),

		NATSURL:      getEnv("NATS_URL", "nats://localhost:4222"),
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", ""),
		DefaultEmail: getEnv("DEFAULT_EMAIL_TO", ""),
	}

	if cfg.DBURL == "" {
		cfg.DBURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require&search_path=public",
			cfg.DBUser,
			cfg.DBPassword,
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBName,
		)
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
