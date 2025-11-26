package model

import "github.com/google/uuid"

type NotificationSent struct {
	NotificationID uuid.UUID
	UserID         uuid.UUID
	Channel        ChannelType
}

func (e NotificationSent) Type() string {
	return "NotificationSent"
}

type NotificationFailed struct {
	UserID  uuid.UUID
	Channel ChannelType
	Reason  string
}

func (e NotificationFailed) Type() string {
	return "NotificationFailed"
}
