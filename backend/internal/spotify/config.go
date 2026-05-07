package spotify

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	RedisURL            string `validate:"required,uri"`
	RabbitMQURL         string `validate:"required,uri"`
	SpotifyClientID     string `validate:"required,min=1"`
	SpotifyClientSecret string `validate:"required,min=1"`
}

func LoadConfig() (*Config, error) {
	godotenv.Load("./.env")

	cfg := &Config{
		RedisURL:            os.Getenv("REDIS_URL"),
		RabbitMQURL:         os.Getenv("RABBITMQ_URL"),
		SpotifyClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		SpotifyClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
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
