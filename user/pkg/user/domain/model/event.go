package model

import (
	"time"

	"github.com/google/uuid"
)

type UserCreated struct {
	UserID    uuid.UUID
	Status    UserStatus
	Login     string
	Email     *string
	Telegram  *string
	CreatedAt time.Time
}

func (u UserCreated) Type() string {
	return "user_created"
}

type UserUpdated struct {
	UserID        uuid.UUID
	UpdatedFields *struct {
		Status   *UserStatus
		Email    *string
		Telegram *string
	}
	RemovedFields *struct {
		Email    *bool
		Telegram *bool
	}
	UpdatedAt time.Time
}

func (u UserUpdated) Type() string {
	return "user_updated"
}

type UserDeleted struct {
	UserID    uuid.UUID
	Status    UserStatus
	DeletedAt time.Time
	Hard      bool
}

func (u UserDeleted) Type() string {
	return "user_deleted"
}
