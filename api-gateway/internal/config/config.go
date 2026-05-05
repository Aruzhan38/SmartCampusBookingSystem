package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	HTTPPort           string
	UserServiceAddr    string
	RoomServiceAddr    string
	BookingServiceAddr string
	JWTSecret          string
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

	return &Config{
		HTTPPort:           firstNonEmpty(os.Getenv("HTTP_PORT"), viper.GetString("http.port")),
		UserServiceAddr:    firstNonEmpty(os.Getenv("USER_SERVICE_ADDR"), viper.GetString("services.user")),
		RoomServiceAddr:    firstNonEmpty(os.Getenv("ROOM_SERVICE_ADDR"), viper.GetString("services.room")),
		BookingServiceAddr: firstNonEmpty(os.Getenv("BOOKING_SERVICE_ADDR"), viper.GetString("services.booking")),
		JWTSecret:          firstNonEmpty(os.Getenv("JWT_SECRET"), viper.GetString("jwt.secret")),
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
