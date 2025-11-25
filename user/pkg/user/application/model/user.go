package model

import "github.com/google/uuid"

type User struct {
	UserID   uuid.UUID
	Status   int
	Login    string
	Email    *string
	Telegram *string
}
