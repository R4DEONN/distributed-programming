package main

import (
	"gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/amqp"
)

func newAMQPConnection(config AMQP, logger logging.Logger) amqp.Connection {
	return amqp.NewAMQPConnection(appID, &amqp.ConnectionConfig{
		User:           config.User,
		Password:       config.Password,
		Host:           config.Host,
		ConnectTimeout: config.ConnectTimeout,
	}, logger)
}
