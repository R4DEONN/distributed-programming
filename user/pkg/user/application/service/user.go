package service

import (
	"context"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
	"github.com/google/uuid"

	"userservice/pkg/common/domain"
	appmodel "userservice/pkg/user/application/model"
	"userservice/pkg/user/domain/model"
	"userservice/pkg/user/domain/service"
)

type UserService interface {
	StoreUser(ctx context.Context, user appmodel.User) (uuid.UUID, error)
	SetUserStatus(ctx context.Context, userID uuid.UUID, status int) error
	FindUser(ctx context.Context, userID uuid.UUID) (appmodel.User, error)
}

func NewUserService(
	uow UnitOfWork,
	luow LockableUnitOfWork,
	eventDispatcher outbox.EventDispatcher[outbox.Event],
) UserService {
	return &userService{
		uow:             uow,
		luow:            luow,
		eventDispatcher: eventDispatcher,
	}
}

type userService struct {
	uow             UnitOfWork
	luow            LockableUnitOfWork
	eventDispatcher outbox.EventDispatcher[outbox.Event]
}

func (s *userService) StoreUser(ctx context.Context, user appmodel.User) (uuid.UUID, error) {
	var lockNames []string
	if user.UserID != uuid.Nil {
		lockNames = append(lockNames, userLock(user.UserID))
	} else {
		lockNames = append(lockNames, userLoginLock(user.Login))
	}
	if user.Email != nil {
		lockNames = append(lockNames, userEmailLock(*user.Email))
	}
	if user.Telegram != nil {
		lockNames = append(lockNames, userTelegramLock(*user.Telegram))
	}

	userID := user.UserID
	err := s.luow.Execute(ctx, lockNames, func(provider RepositoryProvider) error {
		domainService := s.domainService(ctx, provider.UserRepository(ctx))
		if user.UserID == uuid.Nil {
			uID, err := domainService.CreateUser(user.Login)
			if err != nil {
				return err
			}
			userID = uID
		}

		err := domainService.UpdateUserEmail(userID, user.Email)
		if err != nil {
			return err
		}

		err = domainService.UpdateUserTelegram(userID, user.Telegram)
		if err != nil {
			return err
		}

		return nil
	})
	return userID, err
}

func (s *userService) SetUserStatus(ctx context.Context, userID uuid.UUID, status int) error {
	return s.luow.Execute(ctx, []string{userLock(userID)}, func(provider RepositoryProvider) error {
		return s.domainService(ctx, provider.UserRepository(ctx)).UpdateUserStatus(userID, model.UserStatus(status))
	})
}

func (s *userService) FindUser(ctx context.Context, userID uuid.UUID) (appmodel.User, error) {
	var user appmodel.User
	err := s.luow.Execute(ctx, []string{userLock(userID)}, func(provider RepositoryProvider) error {
		domainUser, err := provider.UserRepository(ctx).Find(model.FindSpec{UserID: &userID})
		if err != nil {
			return err
		}
		user = appmodel.User{
			UserID:   domainUser.UserID,
			Status:   int(domainUser.Status),
			Login:    domainUser.Login,
			Email:    domainUser.Email,
			Telegram: domainUser.Telegram,
		}
		return nil
	})
	return user, err
}

func (s *userService) domainService(ctx context.Context, repository model.UserRepository) service.UserService {
	return service.NewUserService(repository, s.domainEventDispatcher(ctx))
}

func (s *userService) domainEventDispatcher(ctx context.Context) domain.EventDispatcher {
	return &domainEventDispatcher{
		ctx:             ctx,
		eventDispatcher: s.eventDispatcher,
	}
}

const baseUserLock = "user_"

func userLock(id uuid.UUID) string {
	return baseUserLock + id.String()
}

func userLoginLock(login string) string {
	return baseUserLock + "login_" + login
}

func userEmailLock(email string) string {
	return baseUserLock + "email_" + email
}

func userTelegramLock(telegram string) string {
	return baseUserLock + "telegram_" + telegram
}
