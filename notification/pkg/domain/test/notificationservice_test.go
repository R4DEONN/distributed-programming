package tests

import (
	"testing"
	"time"

	"notification/pkg/domain/model"
	"notification/pkg/domain/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNotificationService(t *testing.T) {
	t.Run("HandleUserCreated", func(t *testing.T) {
		repo := &mockNotificationRepository{
			recipients: make(map[uuid.UUID]*model.Recipient),
			logs:       make([]*model.NotificationLog, 0),
		}
		dispatcher := &mockEventDispatcher{}
		svc := service.NewNotificationService(repo, dispatcher)

		event := model.UserCreatedEvent{
			UserID:     uuid.New(),
			Email:      "test@test.ru",
			TelegramID: "testIf",
		}

		err := svc.HandleUserCreated(event)
		require.NoError(t, err)

		recipient, err := repo.FindRecipientByUserID(event.UserID)
		require.NoError(t, err)
		require.Equal(t, event.Email, recipient.Email)
		require.Equal(t, event.TelegramID, recipient.TelegramID)
	})

	t.Run("HandleOrderStatusChanged_SendsNotificationForExistingRecipient", func(t *testing.T) {
		repo := &mockNotificationRepository{
			recipients: make(map[uuid.UUID]*model.Recipient),
			logs:       make([]*model.NotificationLog, 0),
		}
		dispatcher := &mockEventDispatcher{}
		svc := service.NewNotificationService(repo, dispatcher)

		userID := uuid.New()
		_ = svc.HandleUserCreated(model.UserCreatedEvent{
			UserID: userID,
			Email:  "test@test.ru",
		})

		orderStatusEvent := model.OrderStatusChangedEvent{
			OrderID:   uuid.New(),
			UserID:    userID,
			NewStatus: "Paid",
		}

		err := svc.HandleOrderStatusChanged(orderStatusEvent)
		require.NoError(t, err)

		require.Len(t, repo.logs, 1)
		require.Equal(t, userID, repo.logs[0].UserID)
		require.Contains(t, repo.logs[0].Message, "Paid")

		require.Len(t, dispatcher.events, 1)
		_, ok := dispatcher.events[0].(model.NotificationSent)
		require.True(t, ok)
	})

	t.Run("HandleOrderStatusChanged_FailsGracefullyForNonExistentRecipient", func(t *testing.T) {
		repo := &mockNotificationRepository{
			recipients: make(map[uuid.UUID]*model.Recipient),
			logs:       make([]*model.NotificationLog, 0),
		}
		dispatcher := &mockEventDispatcher{}
		svc := service.NewNotificationService(repo, dispatcher)

		orderStatusEvent := model.OrderStatusChangedEvent{
			OrderID:   uuid.New(),
			UserID:    uuid.New(),
			NewStatus: "Pending",
		}

		err := svc.HandleOrderStatusChanged(orderStatusEvent)
		require.Error(t, err)
		require.ErrorIs(t, err, model.ErrRecipientNotFound)

		require.Empty(t, repo.logs)
		require.Empty(t, dispatcher.events)
	})
}

var _ model.NotificationRepository = (*mockNotificationRepository)(nil)

type mockNotificationRepository struct {
	recipients map[uuid.UUID]*model.Recipient
	logs       []*model.NotificationLog
}

func (m *mockNotificationRepository) StoreRecipient(r *model.Recipient) error {
	now := time.Now()
	if r.CreatedAt.IsZero() {
		r.CreatedAt = now
	}
	r.UpdatedAt = now
	m.recipients[r.UserID] = r
	return nil
}

func (m *mockNotificationRepository) FindRecipientByUserID(userID uuid.UUID) (*model.Recipient, error) {
	if r, ok := m.recipients[userID]; ok {
		return r, nil
	}
	return nil, model.ErrRecipientNotFound
}

func (m *mockNotificationRepository) StoreLog(log *model.NotificationLog) error {
	log.SentAt = time.Now()
	m.logs = append(m.logs, log)
	return nil
}

var _ service.EventDispatcher = (*mockEventDispatcher)(nil)

type mockEventDispatcher struct {
	events []service.Event
}

func (m *mockEventDispatcher) Dispatch(e service.Event) error {
	m.events = append(m.events, e)
	return nil
}
