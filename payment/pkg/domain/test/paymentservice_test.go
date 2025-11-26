package tests

import (
	"payment/pkg/domain/model"
	"payment/pkg/domain/service"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ model.PaymentRepository = (*mockPaymentRepository)(nil)

type mockPaymentRepository struct {
	accountsByUserID map[uuid.UUID]*model.Account
	txsByOrderID     map[uuid.UUID]*model.Transaction
}

func newMockPaymentRepository() *mockPaymentRepository {
	return &mockPaymentRepository{
		accountsByUserID: make(map[uuid.UUID]*model.Account),
		txsByOrderID:     make(map[uuid.UUID]*model.Transaction),
	}
}
func (m *mockPaymentRepository) NextID() (uuid.UUID, error) { return uuid.NewV7() }
func (m *mockPaymentRepository) StoreAccount(a *model.Account) error {
	m.accountsByUserID[a.UserID] = a
	return nil
}
func (m *mockPaymentRepository) FindAccountByUserID(userID uuid.UUID) (*model.Account, error) {
	if a, ok := m.accountsByUserID[userID]; ok {
		return a, nil
	}
	return nil, model.ErrAccountNotFound
}
func (m *mockPaymentRepository) StoreTransaction(tx *model.Transaction) error {
	if _, ok := m.txsByOrderID[tx.OrderID]; ok {
		return model.ErrDuplicateTransaction
	}
	m.txsByOrderID[tx.OrderID] = tx
	return nil
}
func (m *mockPaymentRepository) FindTransactionByOrderID(orderID uuid.UUID) (*model.Transaction, error) {
	if tx, ok := m.txsByOrderID[orderID]; ok {
		return tx, nil
	}
	return nil, model.ErrDuplicateTransaction
}

var _ service.EventDispatcher = (*mockEventDispatcher)(nil)

type mockEventDispatcher struct{ events []service.Event }

func (m *mockEventDispatcher) Dispatch(e service.Event) error {
	m.events = append(m.events, e)
	return nil
}
func (m *mockEventDispatcher) Clear() { m.events = nil }

func TestPaymentService_ProcessPayment(t *testing.T) {
	repo := newMockPaymentRepository()
	dispatcher := &mockEventDispatcher{}
	paymentService := service.NewPaymentService(repo, dispatcher)

	userID := uuid.New()
	_, _ = paymentService.CreateAccount(userID, 100.0)

	t.Run("successful payment", func(t *testing.T) {
		dispatcher.Clear()
		orderID := uuid.New()

		tx, err := paymentService.ProcessPayment(userID, orderID, 75.0)

		require.NoError(t, err)
		assert.NotNil(t, tx)

		account, _ := repo.FindAccountByUserID(userID)
		assert.Equal(t, 25.0, account.Balance)

		storedTx, _ := repo.FindTransactionByOrderID(orderID)
		assert.Equal(t, tx.ID, storedTx.ID)

		require.Len(t, dispatcher.events, 1)
		_, ok := dispatcher.events[0].(model.PaymentSucceeded)
		assert.True(t, ok)
	})

	t.Run("fails due to insufficient funds", func(t *testing.T) {
		dispatcher.Clear()
		orderID := uuid.New()

		tx, err := paymentService.ProcessPayment(userID, orderID, 50.0)

		require.Error(t, err)
		assert.ErrorIs(t, err, model.ErrInsufficientFunds)
		assert.Nil(t, tx)

		account, _ := repo.FindAccountByUserID(userID)
		assert.Equal(t, 25.0, account.Balance)

		require.Len(t, dispatcher.events, 1)
		event, ok := dispatcher.events[0].(model.PaymentFailed)
		require.True(t, ok)
		assert.Equal(t, "InsufficientFunds", event.Reason)
	})

	t.Run("idempotency check", func(t *testing.T) {
		repo := newMockPaymentRepository()
		dispatcher := &mockEventDispatcher{}
		paymentService := service.NewPaymentService(repo, dispatcher)
		userID := uuid.New()
		_, _ = paymentService.CreateAccount(userID, 200.0)
		orderID := uuid.New()

		tx1, err1 := paymentService.ProcessPayment(userID, orderID, 50.0)
		require.NoError(t, err1)
		assert.Equal(t, 150.0, repo.accountsByUserID[userID].Balance)
		dispatcher.Clear()

		tx2, err2 := paymentService.ProcessPayment(userID, orderID, 50.0)
		require.NoError(t, err2, "Second call should not return an error")
		assert.Equal(t, 150.0, repo.accountsByUserID[userID].Balance, "Balance should not change on second call")
		assert.Equal(t, tx1.ID, tx2.ID, "Should return the original transaction")
		assert.Empty(t, dispatcher.events, "No new event should be dispatched")
	})

	t.Run("fails when account not found", func(t *testing.T) {
		dispatcher.Clear()
		_, err := paymentService.ProcessPayment(uuid.New(), uuid.New(), 50.0)
		assert.ErrorIs(t, err, model.ErrAccountNotFound)
	})
}
