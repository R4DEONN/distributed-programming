package main

import (
	"errors"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	libio "gitea.xscloud.ru/xscloud/golib/pkg/common/io"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	outboxmigrations "gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/outbox/migrations"
	"github.com/urfave/cli/v2"

	"userservice/pkg/user/infrastructure/integrationevent"
	"userservice/pkg/user/infrastructure/migrations/database"
)

type migrateConfig struct {
	Database Database `envconfig:"database" required:"true"`
}

func migrate(logger logging.Logger) *cli.Command {
	return &cli.Command{
		Name: "migrate",
		Subcommands: cli.Commands{
			&cli.Command{
				Name:   "database",
				Action: migrateImpl(logger),
			},
		},
	}
}

func migrateImpl(logger logging.Logger) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		cnf, err := parseEnvs[migrateConfig]()
		if err != nil {
			return err
		}

		closer := libio.NewMultiCloser()
		defer func() {
			err = errors.Join(err, closer.Close())
		}()

		connector, err := newDatabaseConnector(cnf.Database)
		if err != nil {
			return err
		}
		closer.AddCloser(connector)
		connPool := mysql.NewConnectionPool(connector.TransactionalClient())

		databaseMigrator, closeDatabaseMigrator, err := database.NewDatabaseMigrator(c.Context, connPool, logger)
		if err != nil {
			return err
		}
		closer.AddCloser(libio.CloserFunc(closeDatabaseMigrator))

		domainOutboxMigrator, domainOutboxRelease, err := outboxmigrations.NewOutboxMigrator(c.Context, connPool, logger, integrationevent.TransportName)
		if err != nil {
			return err
		}
		closer.AddCloser(domainOutboxRelease)

		err = databaseMigrator.Migrate()
		if err != nil {
			return err
		}
		err = domainOutboxMigrator.Migrate()
		if err != nil {
			return err
		}

		return nil
	}
}
