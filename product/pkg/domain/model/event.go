package model

import "github.com/google/uuid"

type ProductCreated struct {
	ProductID uuid.UUID
	Name      string
	Price     float64
}

func (e ProductCreated) Type() string {
	return "ProductCreated"
}

type ProductUpdated struct {
	ProductID uuid.UUID
	OldName   string
	NewName   string
	OldPrice  float64
	NewPrice  float64
}

func (e ProductUpdated) Type() string {
	return "ProductUpdated"
}

type ProductDeleted struct {
	ProductID uuid.UUID
}

func (e ProductDeleted) Type() string {
	return "ProductDeleted"
}
