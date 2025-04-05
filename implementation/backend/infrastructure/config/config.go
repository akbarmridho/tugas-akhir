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
	TlsCertPath string `envconfig:"TLS_CERT_PATH" required:"true"`
	TlsKeyPath  string `envconfig:"TLS_KEY_PATH" required:"true"`
	JwtSecret   string `envconfig:"JWT_SECRET" required:"true"`

	TestScenario string `envconfig:"TEST_SCENARIO"`
	AppVariant   AppVariant
	PodName      string `envconfig:"POD_NAME" default:"none"`

	DatabaseUrl       string `envconfig:"DATABASE_URL"`
	RisingwaveUrl     string `envconfig:"RISINGWAVE_URL"`
	ScyllaHosts       string `envconfig:"SCYLLA_HOSTS"`
	AmqpUrl           string `envconfig:"AMQP_URL"`
	PaymentServiceUrl string `envconfig:"PAYMENT_SERVICE_URL" required:"true"`
	PaymentCertPath   string `envconfig:"PAYMENT_CERT_PATH" required:"true"`
	WebhookSecret     string `envconfig:"WEBHOOK_SECRET" required:"true"`

	WorkerMetricsPort int `envconfig:"WORKER_METRICS_PORT" default:"5000"`

	RedisHosts    string `envconfig:"REDIS_HOSTS"`
	RedisPassword string `envconfig:"REDIS_PASSWORD"`
}

func NewConfig() (*Config, error) {
	var config Config

	if err := envconfig.Process("", &config); err != nil {
		return nil, errors.Wrap(err, "Missing env variable")
	}

	return &config, nil
}

var Module = fx.Options(fx.Provide(NewConfig))
