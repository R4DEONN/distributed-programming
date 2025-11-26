package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrAccountNotFound      = errors.New("user account not found")
	ErrInsufficientFunds    = errors.New("insufficient funds on account")
	ErrDuplicateTransaction = errors.New("transaction for this order already exists")
	ErrNegativeAmount       = errors.New("amount cannot be negative")
)

type Account struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Balance   float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Transaction struct {
	ID        uuid.UUID
	AccountID uuid.UUID
	OrderID   uuid.UUID
	Amount    float64
	Timestamp time.Time
}

type PaymentRepository interface {
	NextID() (uuid.UUID, error)

	StoreAccount(account *Account) error
	FindAccountByUserID(userID uuid.UUID) (*Account, error)

	StoreTransaction(transaction *Transaction) error
	FindTransactionByOrderID(orderID uuid.UUID) (*Transaction, error)
}
