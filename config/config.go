package config

import (
	"fmt"
	"github.com/start-codex/goevents/rpc"
	"github.com/start-codex/goevents/web"
	"os"
	"time"
)

type (
	SqlConfig struct {
		Conn string `required:"true"`
	}

	NatsConfig struct {
		URL    string `required:"true"`
		Stream string `default:"mallbots"`
	}

	OtelConfig struct {
		ServiceName      string `envconfig:"SERVICE_NAME" default:"service"`
		ExporterEndpoint string `envconfig:"EXPORTER_OTLP_ENDPOINT" default:"http://collector:4317"`
	}

	AppConfig struct {
		Environment     string
		LogLevel        string `envconfig:"LOG_LEVEL" default:"DEBUG"`
		PG              SqlConfig
		Nats            NatsConfig
		Rpc             rpc.RpcConfig
		Web             web.WebConfig
		Otel            OtelConfig
		ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"30s"`
	}
)

func InitConfig() (cfg AppConfig, err error) {
	serviceName := os.Getenv("SERVICE_NAME")
	if len(serviceName) == 0 {
		err = fmt.Errorf("SERVICE_NAME environment variable is not set")
	}

	dbDriver := os.Getenv("DB_DRIVER")
	if len(dbDriver) == 0 {
		err = fmt.Errorf("DB_DRIVER environment variable is not set")
	}

	return
}
