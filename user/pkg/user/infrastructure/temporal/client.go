package temporal

import (
	"errors"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	"go.temporal.io/sdk/client"
)

func NewClient(logger logging.Logger, address string) (client.Client, error) {
	return client.NewLazyClient(client.Options{
		HostPort: address,
		Logger:   &temporalLogger{logger: logger},
	})
}

type temporalLogger struct {
	logger logging.Logger
}

func (l *temporalLogger) Debug(msg string, keyvals ...interface{}) {
	args := append([]interface{}{msg}, keyvals...)
	l.logger.Debug(args...)
}

func (l *temporalLogger) Info(msg string, keyvals ...interface{}) {
	args := append([]interface{}{msg}, keyvals...)
	l.logger.Info(args...)
}

func (l *temporalLogger) Warn(msg string, keyvals ...interface{}) {
	l.logger.Warning(errors.New(msg), keyvals...)
}

func (l *temporalLogger) Error(msg string, keyvals ...interface{}) {
	l.logger.Error(errors.New(msg), keyvals...)
}
