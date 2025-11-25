package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound            = errors.New("user not found")
	ErrUserLoginAlreadyUsed    = errors.New("user login already used")
	ErrUserEmailAlreadyUsed    = errors.New("user email already used")
	ErrUserTelegramAlreadyUsed = errors.New("user telegram already used")
)

type UserStatus int

const (
	Blocked UserStatus = iota
	Active
	Deleted
)

type User struct {
	UserID    uuid.UUID
	Status    UserStatus
	Login     string
	Email     *string
	Telegram  *string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type FindSpec struct {
	UserID   *uuid.UUID
	Login    *string
	Email    *string
	Telegram *string
}

type UserRepository interface {
	NextID() (uuid.UUID, error)
	Store(user User) error
	Find(spec FindSpec) (*User, error)
	HardDelete(userID uuid.UUID) error
}
