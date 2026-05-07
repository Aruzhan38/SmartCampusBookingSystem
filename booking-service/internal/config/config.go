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
		DBName:     getEnv("DB_NAME", "bookings"),
		DBPort:     getEnv("DB_PORT", "5432"),
		GRPCPort:   getEnv("GRPC_PORT", "50053"),
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
