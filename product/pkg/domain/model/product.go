package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrProductNotFound     = errors.New("product not found")
	ErrProductNameExists   = errors.New("product with this name already exists")
	ErrProductNameRequired = errors.New("product name is required")
	ErrProductPriceInvalid = errors.New("product price must be zero or positive")
)

type Product struct {
	ID        uuid.UUID
	Name      string
	Price     float64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type ProductRepository interface {
	NextID() (uuid.UUID, error)
	Store(product *Product) error
	Find(id uuid.UUID) (*Product, error)
	FindByName(name string) (*Product, error)
	Delete(id uuid.UUID) error
	ListAll() ([]*Product, error)
}
