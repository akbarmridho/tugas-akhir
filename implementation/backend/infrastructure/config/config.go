package config

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"go.uber.org/fx"
)

type Config struct {
	Environment string `envconfig:"ENVIRONMENT" required:"true" default:"development"`
	ServerPort  int    `envconfig:"SERVER_PORT" default:"3000"`
	JwtSecret   string `envconfig:"JWT_SECRET" required:"true"`

	DatabaseUrl string `envconfig:"DATABASE_URL" required:"true"`
}

func NewConfig() (*Config, error) {
	var config Config

	if err := envconfig.Process("", &config); err != nil {
		return nil, errors.Wrap(err, "Missing env variable")
	}

	return &config, nil
}

var Module = fx.Options(fx.Provide(NewConfig))
