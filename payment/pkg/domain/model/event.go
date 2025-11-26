package model

import "github.com/google/uuid"

type PaymentSucceeded struct {
	TransactionID uuid.UUID
	OrderID       uuid.UUID
	UserID        uuid.UUID
	Amount        float64
}

func (e PaymentSucceeded) Type() string {
	return "PaymentSucceeded"
}

type PaymentFailed struct {
	OrderID uuid.UUID
	UserID  uuid.UUID
	Reason  string
}

func (e PaymentFailed) Type() string {
	return "PaymentFailed"
}
