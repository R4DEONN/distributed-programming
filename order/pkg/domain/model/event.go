package model

import "github.com/google/uuid"

type OrderCreated struct {
	OrderID    uuid.UUID
	CustomerID uuid.UUID
}

func (e OrderCreated) Type() string {
	return "OrderCreated"
}

type OrderItemChanged struct {
	OrderID      uuid.UUID
	AddedItems   []uuid.UUID
	RemovedItems []uuid.UUID
}

func (e OrderItemChanged) Type() string {
	return "OrderItemChanged"
}

type OrderRemoved struct {
	OrderID uuid.UUID
}

func (e OrderRemoved) Type() string {
	return "OrderRemoved"
}

type OrderStatusChanged struct {
	OrderID uuid.UUID
	Status  OrderStatus
}

func (e OrderStatusChanged) Type() string {
	return "OrderStatusChanged"
}

type OrderItemRemoved struct {
	OrderID uuid.UUID
	ItemID  uuid.UUID
}

func (e OrderItemRemoved) Type() string {
	return "OrderItemRemoved"
}
