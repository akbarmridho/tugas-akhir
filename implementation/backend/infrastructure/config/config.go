package config

import (
	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"go.uber.org/fx"
)

type Config struct {
	Timezone    string `envconfig:"TIMEZONE" default:"Asia/Jakarta"`
	Environment string `envconfig:"ENVIRONMENT" required:"true" default:"development"`

	ServerPort int    `envconfig:"SERVER_PORT" default:"3000"`
	ServerCors string `envconfig:"SERVER_CORS" required:"true" default:"https://localhost:5173"`

	DatabaseHost     string `envconfig:"DATABASE_HOST" required:"true"`
	DatabasePort     string `envconfig:"DATABASE_PORT" required:"true"`
	DatabaseUsername string `envconfig:"DATABASE_USERNAME" required:"true"`
	DatabasePassword string `envconfig:"DATABASE_PASSWORD" required:"true"`
	DatabaseName     string `envconfig:"DATABASE_NAME" required:"true"`
	DatabaseVerifyCA string `envconfig:"DATABASE_VERIFY_CA" required:"false" default:"false"`

	JwtSecret string `envconfig:"JWT_SECRET" required:"true"`

	ServerHost string `envconfig:"SERVER_HOST" required:"true" default:"http://localhost:3000"`
}

func NewConfig() (*Config, error) {
	var config Config

	if err := envconfig.Process("", &config); err != nil {
		return nil, errors.Wrap(err, "Missing env variable")
	}

	return &config, nil
}

var Module = fx.Options(fx.Provide(NewConfig))
