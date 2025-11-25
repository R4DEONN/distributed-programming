package integrationevent

import (
	"encoding/json"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/outbox"
	"github.com/pkg/errors"

	"userservice/pkg/user/domain/model"
)

func NewEventSerializer() outbox.EventSerializer[outbox.Event] {
	return &eventSerializer{}
}

type eventSerializer struct{}

func (s eventSerializer) Serialize(event outbox.Event) (string, error) {
	switch e := event.(type) {
	case *model.UserCreated:
		b, err := json.Marshal(UserCreated{
			UserID:    e.UserID.String(),
			Status:    int(e.Status),
			Login:     e.Login,
			Email:     e.Email,
			Telegram:  e.Telegram,
			CreatedAt: e.CreatedAt.Unix(),
		})
		return string(b), errors.WithStack(err)
	case *model.UserUpdated:
		ie := UserUpdated{
			UserID:    e.UserID.String(),
			UpdatedAt: e.UpdatedAt.Unix(),
		}
		if e.UpdatedFields != nil {
			ie.UpdatedFields = &struct {
				Status   *int    `json:"status,omitempty"`
				Email    *string `json:"email,omitempty"`
				Telegram *string `json:"telegram,omitempty"`
			}{
				Status:   (*int)(e.UpdatedFields.Status),
				Email:    e.UpdatedFields.Email,
				Telegram: e.UpdatedFields.Telegram,
			}
		}
		if e.RemovedFields != nil {
			ie.RemovedFields = &struct {
				Email    *bool `json:"email,omitempty"`
				Telegram *bool `json:"telegram,omitempty"`
			}{
				Email:    e.RemovedFields.Email,
				Telegram: e.RemovedFields.Telegram,
			}
		}
		b, err := json.Marshal(ie)
		return string(b), errors.WithStack(err)
	case *model.UserDeleted:
		b, err := json.Marshal(UserDeleted{
			UserID:    e.UserID.String(),
			Status:    int(e.Status),
			DeletedAt: e.DeletedAt.Unix(),
			Hard:      e.Hard,
		})
		return string(b), errors.WithStack(err)
	default:
		return "", errors.Errorf("unknown event %q", event.Type())
	}
}

type UserCreated struct {
	UserID    string  `json:"user_id"`
	Status    int     `json:"status"`
	Login     string  `json:"login"`
	Email     *string `json:"email,omitempty"`
	Telegram  *string `json:"telegram,omitempty"`
	CreatedAt int64   `json:"created_at"`
}

type UserUpdated struct {
	UserID        string `json:"user_id"`
	UpdatedFields *struct {
		Status   *int    `json:"status,omitempty"`
		Email    *string `json:"email,omitempty"`
		Telegram *string `json:"telegram,omitempty"`
	} `json:"updated_fields,omitempty"`
	RemovedFields *struct {
		Email    *bool `json:"email,omitempty"`
		Telegram *bool `json:"telegram,omitempty"`
	} `json:"removed_fields,omitempty"`
	UpdatedAt int64 `json:"updated_at,omitempty"`
}

type UserDeleted struct {
	UserID    string `json:"user_id"`
	Status    int    `json:"status"`
	DeletedAt int64  `json:"deleted_at"`
	Hard      bool   `json:"hard"`
}
