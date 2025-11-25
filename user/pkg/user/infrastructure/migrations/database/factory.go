package database

import (
	"context"
	"errors"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	libmigrator "gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/migrator"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
)

type MigrationBuilderFunc func(client mysql.ClientContext) libmigrator.Migration
type ReleaseConnectionFunc func() error

func NewDatabaseMigrator(
	ctx context.Context,
	pool mysql.ConnectionPool,
	logger logging.Logger,
) (migrator libmigrator.Migrator, release ReleaseConnectionFunc, err error) {
	conn, err2 := pool.TransactionalConnection(ctx)
	if err2 != nil {
		return nil, nil, err2
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, conn.Close())
		}
	}()

	l := logger.WithField("migrator", "database")
	factory := libmigrator.NewMigratorFactory("database", conn, l)

	migrations := make([]libmigrator.Migration, 0, len(builderFunctions))
	for _, builder := range builderFunctions {
		migrations = append(migrations, builder(conn))
	}

	migrator, err = factory.NewMigrator(ctx, migrations...)
	if err != nil {
		return nil, nil, err
	}
	return migrator, conn.Close, nil
}

var builderFunctions = []MigrationBuilderFunc{
	NewVersion1722266003,
}
