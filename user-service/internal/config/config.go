package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	DBURL       string
	DatabaseURL string
	DBHost      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBPort      string
	JWTSecret   string
	GRPCPort    string
}

func Load() *Config {
	_ = godotenv.Load()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := viper.ReadInConfig(); err != nil {
		// No config file is fine when using environment variables.
	}

	dbURL := firstNonEmpty(
		os.Getenv("DB_URL"),
		os.Getenv("DATABASE_URL"),
		viper.GetString("db.url"),
		viper.GetString("database_url"),
	)

	if dbURL == "" {
		dbURL = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=require channel_binding=require",
			viper.GetString("db.host"),
			viper.GetString("db.user"),
			viper.GetString("db.password"),
			viper.GetString("db.name"),
			viper.GetString("db.port"),
		)
	}

	return &Config{
		DBURL:       dbURL,
		DatabaseURL: dbURL,
		DBHost:      viper.GetString("db.host"),
		DBUser:      viper.GetString("db.user"),
		DBPassword:  viper.GetString("db.password"),
		DBName:      viper.GetString("db.name"),
		DBPort:      viper.GetString("db.port"),
		JWTSecret:   viper.GetString("jwt.secret"),
		GRPCPort:    viper.GetString("grpc.port"),
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
