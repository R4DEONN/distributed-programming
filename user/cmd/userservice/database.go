package main

import (
	"fmt"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/pkg/errors"
)

func newDatabaseConnector(config Database) (mysql.Connector, error) {
	connector := mysql.NewConnector()
	err := connector.Open(dsn(config), mysql.Config{
		MaxConnections:        config.MaxConnections,
		ConnectionMaxLifeTime: config.ConnectionMaxLifeTime,
		ConnectionMaxIdleTime: config.ConnectionMaxIdleTime,
	})
	return connector, errors.WithStack(err)
}

func dsn(config Database) string {
	return fmt.Sprintf(
		"%s:%s@(%s)/%s?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=true",
		config.User,
		config.Password,
		config.Host,
		config.Name,
	)
}
