package mysql

import (
	"context"
	"time"

	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"

	"userservice/pkg/user/application/service"
)

func NewUnitOfWork(uow mysql.UnitOfWorkWithRepositoryProvider[service.RepositoryProvider]) service.UnitOfWork {
	return &unitOfWork{
		uow: uow,
	}
}

type unitOfWork struct {
	uow mysql.UnitOfWorkWithRepositoryProvider[service.RepositoryProvider]
}

func (u *unitOfWork) Execute(ctx context.Context, f func(provider service.RepositoryProvider) error) error {
	return u.uow.ExecuteWithRepositoryProvider(ctx, f)
}

func NewLockableUnitOfWork(uow mysql.LockableUnitOfWorkWithRepositoryProvider[service.RepositoryProvider]) service.LockableUnitOfWork {
	return &lockableUnitOfWork{
		uow: uow,
	}
}

type lockableUnitOfWork struct {
	uow mysql.LockableUnitOfWorkWithRepositoryProvider[service.RepositoryProvider]
}

func (l *lockableUnitOfWork) Execute(ctx context.Context, lockNames []string, f func(provider service.RepositoryProvider) error) error {
	if len(lockNames) == 1 {
		return l.uow.ExecuteWithRepositoryProvider(ctx, lockNames[0], time.Minute, f)
	}
	ln := lockNames[0]
	lns := lockNames[1:]
	return l.uow.ExecuteWithRepositoryProvider(ctx, ln, time.Minute, func(_ service.RepositoryProvider) error {
		return l.Execute(ctx, lns, f)
	})
}
