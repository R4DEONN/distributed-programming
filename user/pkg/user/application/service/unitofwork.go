package service

import (
	"context"

	"userservice/pkg/user/domain/model"
)

type RepositoryProvider interface {
	UserRepository(ctx context.Context) model.UserRepository
}

type LockableUnitOfWork interface {
	Execute(ctx context.Context, lockNames []string, f func(provider RepositoryProvider) error) error
}
type UnitOfWork interface {
	Execute(ctx context.Context, f func(provider RepositoryProvider) error) error
}
