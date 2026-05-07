package scanner

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string `validate:"required,uri"`
	RedisURL    string `validate:"required,uri"`
	RabbitMQURL string `validate:"required,uri"`
}

func LoadConfig() (*Config, error) {
	godotenv.Load("./.env")

	cfg := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisURL:    os.Getenv("REDIS_URL"),
		RabbitMQURL: os.Getenv("RABBITMQ_URL"),
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
