package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrRecipientNotFound = errors.New("recipient not found")
)

type ChannelType string

const (
	ChannelEmail    ChannelType = "email"
	ChannelTelegram ChannelType = "telegram"
)

type Recipient struct {
	UserID     uuid.UUID
	Email      string
	TelegramID string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type NotificationLog struct {
	ID      uuid.UUID
	UserID  uuid.UUID
	Channel ChannelType
	Message string
	SentAt  time.Time
}

type NotificationRepository interface {
	StoreRecipient(recipient *Recipient) error
	FindRecipientByUserID(userID uuid.UUID) (*Recipient, error)
	StoreLog(log *NotificationLog) error
}
