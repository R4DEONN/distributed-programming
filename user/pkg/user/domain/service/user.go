package service

import (
	"errors"
	"reflect"
	"time"

	"github.com/google/uuid"

	"userservice/pkg/common/domain"
	"userservice/pkg/user/domain/model"
)

type UserService interface {
	CreateUser(login string) (uuid.UUID, error)
	UpdateUserStatus(userID uuid.UUID, status model.UserStatus) error
	UpdateUserEmail(userID uuid.UUID, email *string) error
	UpdateUserTelegram(userID uuid.UUID, telegram *string) error
	DeleteUser(userID uuid.UUID, hard bool) error
}

func NewUserService(
	userRepository model.UserRepository,
	eventDispatcher domain.EventDispatcher,
) UserService {
	return &userService{
		userRepository:  userRepository,
		eventDispatcher: eventDispatcher,
	}
}

type userService struct {
	userRepository  model.UserRepository
	eventDispatcher domain.EventDispatcher
}

func (u userService) CreateUser(login string) (uuid.UUID, error) {
	_, err := u.userRepository.Find(model.FindSpec{
		Login: &login,
	})
	if err != nil && !errors.Is(err, model.ErrUserNotFound) {
		return uuid.Nil, err
	}
	if err == nil {
		return uuid.Nil, model.ErrUserLoginAlreadyUsed
	}

	userID, err := u.userRepository.NextID()
	if err != nil {
		return uuid.Nil, err
	}

	status := model.Blocked
	currentTime := time.Now()
	err = u.userRepository.Store(model.User{
		UserID:    userID,
		Status:    status,
		Login:     login,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	})
	if err != nil {
		return uuid.Nil, err
	}

	return userID, u.eventDispatcher.Dispatch(&model.UserCreated{
		UserID:    userID,
		Status:    status,
		Login:     login,
		CreatedAt: currentTime,
	})
}

func (u userService) UpdateUserStatus(userID uuid.UUID, status model.UserStatus) error {
	user, err := u.userRepository.Find(model.FindSpec{
		UserID: &userID,
	})
	if err != nil {
		return err
	}

	if user.Status == status {
		return nil
	}

	currentTime := time.Now()
	user.Status = status
	user.UpdatedAt = currentTime
	err = u.userRepository.Store(*user)
	if err != nil {
		return err
	}

	return u.eventDispatcher.Dispatch(&model.UserUpdated{
		UserID:    userID,
		UpdatedAt: currentTime,
		UpdatedFields: &struct {
			Status   *model.UserStatus
			Email    *string
			Telegram *string
		}{Status: &status},
	})
}

// nolint:dupl
func (u userService) UpdateUserEmail(userID uuid.UUID, email *string) error {
	user, err := u.userRepository.Find(model.FindSpec{
		UserID: &userID,
	})
	if err != nil {
		return err
	}
	if reflect.DeepEqual(user.Email, email) {
		return nil
	}

	if email != nil {
		userWithEmail, err := u.userRepository.Find(model.FindSpec{
			Email: email,
		})
		if err != nil && !errors.Is(err, model.ErrUserNotFound) {
			return err
		}
		if userWithEmail != nil && userWithEmail.UserID != user.UserID {
			return model.ErrUserEmailAlreadyUsed
		}
	}

	currentTime := time.Now()
	user.Email = email
	user.UpdatedAt = currentTime
	err = u.userRepository.Store(*user)
	if err != nil {
		return err
	}

	if email == nil {
		return u.eventDispatcher.Dispatch(&model.UserUpdated{
			UserID:    userID,
			UpdatedAt: currentTime,
			RemovedFields: &struct {
				Email    *bool
				Telegram *bool
			}{Email: toPtr(true)},
		})
	}

	return u.eventDispatcher.Dispatch(&model.UserUpdated{
		UserID:    userID,
		UpdatedAt: currentTime,
		UpdatedFields: &struct {
			Status   *model.UserStatus
			Email    *string
			Telegram *string
		}{Email: email},
	})
}

// nolint:dupl
func (u userService) UpdateUserTelegram(userID uuid.UUID, telegram *string) error {
	user, err := u.userRepository.Find(model.FindSpec{
		UserID: &userID,
	})
	if err != nil {
		return err
	}
	if reflect.DeepEqual(user.Telegram, telegram) {
		return nil
	}

	if telegram != nil {
		userWithTelegram, err := u.userRepository.Find(model.FindSpec{
			Telegram: telegram,
		})
		if err != nil && !errors.Is(err, model.ErrUserNotFound) {
			return err
		}
		if userWithTelegram != nil && userWithTelegram.UserID != user.UserID {
			return model.ErrUserTelegramAlreadyUsed
		}
	}

	currentTime := time.Now()
	user.Telegram = telegram
	user.UpdatedAt = currentTime
	err = u.userRepository.Store(*user)
	if err != nil {
		return err
	}

	if telegram == nil {
		return u.eventDispatcher.Dispatch(&model.UserUpdated{
			UserID:    userID,
			UpdatedAt: currentTime,
			RemovedFields: &struct {
				Email    *bool
				Telegram *bool
			}{Telegram: toPtr(true)},
		})
	}

	return u.eventDispatcher.Dispatch(&model.UserUpdated{
		UserID:    userID,
		UpdatedAt: currentTime,
		UpdatedFields: &struct {
			Status   *model.UserStatus
			Email    *string
			Telegram *string
		}{Telegram: telegram},
	})
}

func (u userService) DeleteUser(userID uuid.UUID, hard bool) error {
	user, err := u.userRepository.Find(model.FindSpec{
		UserID: &userID,
	})
	if err != nil {
		return err
	}

	if hard {
		err = u.userRepository.HardDelete(userID)
		if err != nil {
			return err
		}
	}

	currentTime := time.Now()
	user.Status = model.Deleted
	user.UpdatedAt = currentTime
	user.DeletedAt = &currentTime
	err = u.userRepository.Store(*user)
	if err != nil {
		return err
	}

	return u.eventDispatcher.Dispatch(&model.UserDeleted{
		UserID:    userID,
		Status:    model.Deleted,
		DeletedAt: currentTime,
		Hard:      hard,
	})
}

func toPtr[T any](v T) *T {
	return &v
}
