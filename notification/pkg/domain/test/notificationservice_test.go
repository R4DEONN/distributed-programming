package tests

import (
	"testing"
	"time"

	"notification/pkg/domain/model"
	"notification/pkg/domain/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ model.NotificationRepository = (*mockNotificationRepository)(nil)

type mockNotificationRepository struct {
	recipients map[uuid.UUID]*model.Recipient
	logs       []*model.NotificationLog
}

func newMockNotificationRepository() *mockNotificationRepository {
	return &mockNotificationRepository{
		recipients: make(map[uuid.UUID]*model.Recipient),
		logs:       make([]*model.NotificationLog, 0),
	}
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

type mockEventDispatcher struct{ events []service.Event }

func (m *mockEventDispatcher) Dispatch(e service.Event) error {
	m.events = append(m.events, e)
	return nil
}
func (m *mockEventDispatcher) Clear() { m.events = nil }

func TestNotificationService_HandleUserCreated(t *testing.T) {
	repo := newMockNotificationRepository()
	notificationService := service.NewNotificationService(repo, &mockEventDispatcher{})

	event := model.UserCreatedEvent{
		UserID:     uuid.New(),
		Email:      "test@example.com",
		TelegramID: "12345",
	}

	err := notificationService.HandleUserCreated(event)
	require.NoError(t, err)

	recipient, err := repo.FindRecipientByUserID(event.UserID)
	require.NoError(t, err)
	assert.Equal(t, event.Email, recipient.Email)
	assert.Equal(t, event.TelegramID, recipient.TelegramID)
}

func TestNotificationService_HandleOrderStatusChanged(t *testing.T) {
	t.Run("sends notification for existing recipient", func(t *testing.T) {
		repo := newMockNotificationRepository()
		dispatcher := &mockEventDispatcher{}
		notificationService := service.NewNotificationService(repo, dispatcher)

		userCreatedEvent := model.UserCreatedEvent{
			UserID: uuid.New(),
			Email:  "notify@me.com",
		}
		_ = notificationService.HandleUserCreated(userCreatedEvent)

		orderStatusEvent := model.OrderStatusChangedEvent{
			OrderID:   uuid.New(),
			UserID:    userCreatedEvent.UserID,
			NewStatus: "PAID",
		}
		err := notificationService.HandleOrderStatusChanged(orderStatusEvent)
		require.NoError(t, err)

		require.Len(t, repo.logs, 1)
		assert.Equal(t, userCreatedEvent.UserID, repo.logs[0].UserID)
		assert.Contains(t, repo.logs[0].Message, "PAID")

		require.Len(t, dispatcher.events, 1)
		_, ok := dispatcher.events[0].(model.NotificationSent)
		assert.True(t, ok)
	})

	t.Run("fails gracefully for non-existent recipient", func(t *testing.T) {
		repo := newMockNotificationRepository()
		dispatcher := &mockEventDispatcher{}
		notificationService := service.NewNotificationService(repo, dispatcher)

		orderStatusEvent := model.OrderStatusChangedEvent{
			OrderID:   uuid.New(),
			UserID:    uuid.New(),
			NewStatus: "SHIPPED",
		}
		err := notificationService.HandleOrderStatusChanged(orderStatusEvent)

		require.Error(t, err)
		assert.ErrorIs(t, err, model.ErrRecipientNotFound)

		assert.Empty(t, repo.logs)
		assert.Empty(t, dispatcher.events)
	})
}
