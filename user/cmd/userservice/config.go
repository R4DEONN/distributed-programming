package main

import (
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

func parseEnvs[T any]() (T, error) {
	var c T
	err := envconfig.Process(appID, &c)
	return c, errors.WithStack(err)
}

type Service struct {
	GracePeriod time.Duration `envconfig:"grace_period" default:"15s"`

	GRPCAddress string `envconfig:"grpc_address" default:":8081"`
	HTTPAddress string `envconfig:"http_address" default:":8082"`
}

type Database struct {
	User                  string        `envconfig:"user" required:"true"`
	Password              string        `envconfig:"password" required:"true"`
	Host                  string        `envconfig:"host" required:"true"`
	Name                  string        `envconfig:"name" required:"true"`
	MaxConnections        int           `envconfig:"max_connections" default:"20"`
	ConnectionMaxLifeTime time.Duration `envconfig:"connection_max_life_time" default:"10m"`
	ConnectionMaxIdleTime time.Duration `envconfig:"connection_max_idle_time" default:"1m"`
}

type AMQP struct {
	User           string        `envconfig:"user" required:"true"`
	Password       string        `envconfig:"password" required:"true"`
	Host           string        `envconfig:"host" required:"true"`
	ConnectTimeout time.Duration `envconfig:"connect_timeout"`
}

type Temporal struct {
	Host string `envconfig:"host" required:"true"`
}
