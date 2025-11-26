package service

import (
	"errors"
	"fmt"
	"time"

	"payment/pkg/domain/model"

	"github.com/google/uuid"
)

type Event interface {
	Type() string
}

type EventDispatcher interface {
	Dispatch(event Event) error
}

type Payment interface {
	CreateAccount(userID uuid.UUID, initialBalance float64) (*model.Account, error)
	ProcessPayment(userID, orderID uuid.UUID, amount float64) (*model.Transaction, error)
	GetAccountByUserID(userID uuid.UUID) (*model.Account, error)
}

func NewPaymentService(repo model.PaymentRepository, dispatcher EventDispatcher) Payment {
	return &paymentService{
		repo:       repo,
		dispatcher: dispatcher,
	}
}

type paymentService struct {
	repo       model.PaymentRepository
	dispatcher EventDispatcher
}

func (s *paymentService) CreateAccount(userID uuid.UUID, initialBalance float64) (*model.Account, error) {
	if initialBalance < 0 {
		return nil, model.ErrNegativeAmount
	}
	id, err := s.repo.NextID()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	account := &model.Account{
		ID:        id,
		UserID:    userID,
		Balance:   initialBalance,
		CreatedAt: now,
		UpdatedAt: now,
	}
	return account, s.repo.StoreAccount(account)
}

func (s *paymentService) ProcessPayment(userID, orderID uuid.UUID, amount float64) (*model.Transaction, error) {
	if amount <= 0 {
		return nil, model.ErrNegativeAmount
	}

	if tx, err := s.repo.FindTransactionByOrderID(orderID); !errors.Is(err, model.ErrDuplicateTransaction) {
		if err == nil {
			return tx, nil
		}
		return nil, fmt.Errorf("failed to check for existing transaction: %w", err)
	}

	account, err := s.repo.FindAccountByUserID(userID)
	if err != nil {
		return nil, err
	}

	if account.Balance < amount {
		_ = s.dispatcher.Dispatch(model.PaymentFailed{
			OrderID: orderID,
			UserID:  userID,
			Reason:  "InsufficientFunds",
		})
		return nil, model.ErrInsufficientFunds
	}

	account.Balance -= amount
	account.UpdatedAt = time.Now()

	txID, err := s.repo.NextID()
	if err != nil {
		return nil, err
	}

	transaction := &model.Transaction{
		ID:        txID,
		AccountID: account.ID,
		OrderID:   orderID,
		Amount:    amount,
		Timestamp: time.Now(),
	}

	if err := s.repo.StoreAccount(account); err != nil {
		return nil, fmt.Errorf("failed to update account balance: %w", err)
	}
	if err := s.repo.StoreTransaction(transaction); err != nil {
		return nil, fmt.Errorf("failed to store transaction: %w", err)
	}

	_ = s.dispatcher.Dispatch(model.PaymentSucceeded{
		TransactionID: transaction.ID,
		OrderID:       orderID,
		UserID:        userID,
		Amount:        amount,
	})

	return transaction, nil
}

func (s *paymentService) GetAccountByUserID(userID uuid.UUID) (*model.Account, error) {
	return s.repo.FindAccountByUserID(userID)
}
