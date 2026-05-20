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
	RedisAddr  string
}

func Load() *Config {
	_ = godotenv.Load()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	_ = viper.ReadInConfig()

	dbURL := firstNonEmpty(
		os.Getenv("DB_URL"),
		os.Getenv("DATABASE_URL"),
		viper.GetString("db.url"),
	)

	if dbURL == "" {
		dbURL = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			firstNonEmpty(os.Getenv("DB_HOST"), viper.GetString("db.host")),
			firstNonEmpty(os.Getenv("DB_USER"), viper.GetString("db.user")),
			firstNonEmpty(os.Getenv("DB_PASSWORD"), viper.GetString("db.password")),
			firstNonEmpty(os.Getenv("DB_NAME"), viper.GetString("db.name")),
			firstNonEmpty(os.Getenv("DB_PORT"), viper.GetString("db.port")),
		)
	}

	return &Config{
		DBURL:    dbURL,
		GRPCPort: firstNonEmpty(os.Getenv("GRPC_PORT"), viper.GetString("grpc.port"), "50052"),
		RedisAddr: firstNonEmpty(
			os.Getenv("REDIS_ADDR"),
			viper.GetString("redis.addr"),
			"localhost:6379",
		),
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
