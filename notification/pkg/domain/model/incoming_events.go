package model

import "github.com/google/uuid"

type UserCreatedEvent struct {
	UserID     uuid.UUID
	Email      string
	TelegramID string
}

type OrderStatusChangedEvent struct {
	OrderID   uuid.UUID
	UserID    uuid.UUID
	NewStatus string
}
