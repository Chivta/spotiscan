package api

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL         string `validate:"required,uri"`
	RedisURL            string `validate:"required,uri"`
	RabbitMQURL         string `validate:"required,uri"`
	JWTSecret           string `validate:"required,min=32"`
	SpotifyClientID     string `validate:"required,min=1"`
	SpotifyClientSecret string `validate:"required,min=1"`
	SecureCookies       bool
}

func LoadConfig() (*Config, error) {
	godotenv.Load("./.env")

	cfg := &Config{
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		RedisURL:            os.Getenv("REDIS_URL"),
		RabbitMQURL:         os.Getenv("RABBITMQ_URL"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		SpotifyClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		SpotifyClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		SecureCookies:       func() bool { v, _ := strconv.ParseBool(os.Getenv("SECURE_COOKIES")); return v }(),
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		var msgs []string
		for _, e := range err.(validator.ValidationErrors) {
			msgs = append(msgs, fmt.Sprintf("field '%s' failed validation: %s", e.Field(), e.Tag()))
		}
		return nil, fmt.Errorf("config validation failed: %v", msgs)
	}

	return cfg, nil
}
