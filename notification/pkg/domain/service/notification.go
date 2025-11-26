package service

import (
	"fmt"

	"notification/pkg/domain/model"

	"github.com/google/uuid"
)

type Event interface {
	Type() string
}

type EventDispatcher interface {
	Dispatch(event Event) error
}

type Notification interface {
	HandleUserCreated(event model.UserCreatedEvent) error
	HandleOrderStatusChanged(event model.OrderStatusChangedEvent) error
}

func NewNotificationService(repo model.NotificationRepository, dispatcher EventDispatcher) Notification {
	return &notificationService{
		repo:       repo,
		dispatcher: dispatcher,
	}
}

type notificationService struct {
	repo       model.NotificationRepository
	dispatcher EventDispatcher
}

func (s *notificationService) HandleUserCreated(event model.UserCreatedEvent) error {
	recipient := &model.Recipient{
		UserID:     event.UserID,
		Email:      event.Email,
		TelegramID: event.TelegramID,
	}
	return s.repo.StoreRecipient(recipient)
}

func (s *notificationService) HandleOrderStatusChanged(event model.OrderStatusChangedEvent) error {
	recipient, err := s.repo.FindRecipientByUserID(event.UserID)
	if err != nil {
		fmt.Printf("error: recipient not found for user %s, cannot send notification\n", event.UserID)
		return err
	}

	message := fmt.Sprintf("Hello! Status of your order %s has changed to: %s", event.OrderID, event.NewStatus)

	logEntry := &model.NotificationLog{
		ID:      uuid.Must(uuid.NewV7()),
		UserID:  recipient.UserID,
		Channel: model.ChannelEmail,
		Message: message,
	}
	if err := s.repo.StoreLog(logEntry); err != nil {
		_ = s.dispatcher.Dispatch(model.NotificationFailed{
			UserID: recipient.UserID, Channel: model.ChannelEmail, Reason: "failed to store log",
		})
		return fmt.Errorf("failed to store notification log: %w", err)
	}

	_ = s.dispatcher.Dispatch(model.NotificationSent{
		NotificationID: logEntry.ID,
		UserID:         recipient.UserID,
		Channel:        model.ChannelEmail,
	})

	fmt.Printf("Sent notification to %s: %s\n", recipient.Email, message)
	return nil
}
