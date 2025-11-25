package tests

import (
	"testing"
	"time"

	"order/pkg/domain/model"
	"order/pkg/domain/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestOrderService(t *testing.T) {
	t.Run("CreateOrder", func(t *testing.T) {
		repo := &mockOrderRepository{store: make(map[uuid.UUID]*model.Order)}
		dispatcher := &mockEventDispatcher{}
		svc := service.NewOrderService(repo, dispatcher)

		customerID := uuid.Must(uuid.NewV7())
		orderID, err := svc.CreateOrder(customerID)
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, orderID)

		stored, err := repo.Find(orderID)
		require.NoError(t, err)
		require.Equal(t, customerID, stored.CustomerID)
		require.Equal(t, model.Open, stored.Status)

		require.Len(t, dispatcher.events, 1)
		event, ok := dispatcher.events[0].(model.OrderCreated)
		require.True(t, ok)
		require.Equal(t, orderID, event.OrderID)
		require.Equal(t, customerID, event.CustomerID)
	})

	t.Run("DeleteOrder", func(t *testing.T) {
		repo := &mockOrderRepository{store: make(map[uuid.UUID]*model.Order)}
		dispatcher := &mockEventDispatcher{}
		svc := service.NewOrderService(repo, dispatcher)

		customerID := uuid.Must(uuid.NewV7())
		orderID, err := svc.CreateOrder(customerID)
		require.NoError(t, err)

		err = svc.DeleteOrder(orderID)
		require.NoError(t, err)

		_, err = repo.Find(orderID)
		require.ErrorIs(t, err, model.ErrOrderNotFound)

		require.NotNil(t, repo.store[orderID].DeletedAt)

		require.Len(t, dispatcher.events, 2)
		event, ok := dispatcher.events[1].(model.OrderRemoved)
		require.True(t, ok)
		require.Equal(t, orderID, event.OrderID)
	})

	t.Run("SetStatus", func(t *testing.T) {
		repo := &mockOrderRepository{store: make(map[uuid.UUID]*model.Order)}
		dispatcher := &mockEventDispatcher{}
		svc := service.NewOrderService(repo, dispatcher)

		customerID := uuid.Must(uuid.NewV7())
		orderID, err := svc.CreateOrder(customerID)
		require.NoError(t, err)

		newStatus := model.Pending
		err = svc.SetStatus(orderID, newStatus)
		require.NoError(t, err)

		stored, err := repo.Find(orderID)
		require.NoError(t, err)
		require.Equal(t, newStatus, stored.Status)

		require.Len(t, dispatcher.events, 2)
		event, ok := dispatcher.events[1].(model.OrderStatusChanged)
		require.True(t, ok)
		require.Equal(t, orderID, event.OrderID)
		require.Equal(t, newStatus, event.Status)
	})

	t.Run("AddItem", func(t *testing.T) {
		repo := &mockOrderRepository{store: make(map[uuid.UUID]*model.Order)}
		dispatcher := &mockEventDispatcher{}
		svc := service.NewOrderService(repo, dispatcher)

		customerID := uuid.Must(uuid.NewV7())
		orderID, err := svc.CreateOrder(customerID)
		require.NoError(t, err)

		productID := uuid.Must(uuid.NewV7())
		price := 199.99

		itemID, err := svc.AddItem(orderID, productID, price)
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, itemID)

		stored, err := repo.Find(orderID)
		require.NoError(t, err)
		require.Len(t, stored.Items, 1)
		item := stored.Items[0]
		require.Equal(t, itemID, item.ID)
		require.Equal(t, productID, item.ProductID)
		require.Equal(t, price, item.Price)

		require.Len(t, dispatcher.events, 2)
		event, ok := dispatcher.events[1].(model.OrderItemChanged)
		require.True(t, ok)
		require.Equal(t, orderID, event.OrderID)
		require.Equal(t, []uuid.UUID{itemID}, event.AddedItems)
	})

	t.Run("AddItem_FailsWhenOrderNotOpen", func(t *testing.T) {
		repo := &mockOrderRepository{store: make(map[uuid.UUID]*model.Order)}
		dispatcher := &mockEventDispatcher{}
		svc := service.NewOrderService(repo, dispatcher)

		customerID := uuid.Must(uuid.NewV7())
		orderID, err := svc.CreateOrder(customerID)
		require.NoError(t, err)

		err = svc.SetStatus(orderID, model.Cancelled)
		require.NoError(t, err)

		productID := uuid.Must(uuid.NewV7())
		_, err = svc.AddItem(orderID, productID, 99.99)
		require.ErrorIs(t, err, service.ErrInvalidOrderStatus)
		require.Len(t, dispatcher.events, 2)
	})

	t.Run("DeleteItem", func(t *testing.T) {
		repo := &mockOrderRepository{store: make(map[uuid.UUID]*model.Order)}
		dispatcher := &mockEventDispatcher{}
		svc := service.NewOrderService(repo, dispatcher)

		customerID := uuid.Must(uuid.NewV7())
		orderID, err := svc.CreateOrder(customerID)
		require.NoError(t, err)

		productID := uuid.Must(uuid.NewV7())
		itemID, err := svc.AddItem(orderID, productID, 50.0)
		require.NoError(t, err)

		err = svc.DeleteItem(orderID, itemID)
		require.NoError(t, err)

		stored, err := repo.Find(orderID)
		require.NoError(t, err)
		require.Len(t, stored.Items, 0)

		require.Len(t, dispatcher.events, 3)
		event, ok := dispatcher.events[2].(model.OrderItemRemoved)
		require.True(t, ok)
		require.Equal(t, orderID, event.OrderID)
		require.Equal(t, itemID, event.ItemID)
	})
}

var _ model.OrderRepository = &mockOrderRepository{}

type mockOrderRepository struct {
	store map[uuid.UUID]*model.Order
}

func (m *mockOrderRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (m *mockOrderRepository) Store(order *model.Order) error {
	m.store[order.ID] = order
	return nil
}

func (m *mockOrderRepository) Find(id uuid.UUID) (*model.Order, error) {
	if order, ok := m.store[id]; ok && order.DeletedAt == nil {
		return order, nil
	}
	return nil, model.ErrOrderNotFound
}

func (m *mockOrderRepository) Delete(id uuid.UUID) error {
	if order, ok := m.store[id]; ok && order.DeletedAt == nil {
		order.DeletedAt = toPtr(time.Now())
		return nil
	}
	return model.ErrOrderNotFound
}

var _ service.EventDispatcher = &mockEventDispatcher{}

type mockEventDispatcher struct {
	events []service.Event
}

func (m *mockEventDispatcher) Dispatch(event service.Event) error {
	m.events = append(m.events, event)
	return nil
}

func toPtr[V any](v V) *V {
	return &v
}
