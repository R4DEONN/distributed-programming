package service

import (
	"errors"
	"fmt"
	"time"

	"product/pkg/domain/model"

	"github.com/google/uuid"
)

type Event interface{ Type() string }
type EventDispatcher interface{ Dispatch(event Event) error }

type Product interface {
	CreateProduct(name string, price float64) (*model.Product, error)
	UpdateProduct(id uuid.UUID, name string, price float64) (*model.Product, error)
	DeleteProduct(id uuid.UUID) error
	GetProduct(id uuid.UUID) (*model.Product, error)
	ListAllProducts() ([]*model.Product, error)
}

func NewProductService(repo model.ProductRepository, dispatcher EventDispatcher) Product {
	return &productService{
		repo:       repo,
		dispatcher: dispatcher,
	}
}

type productService struct {
	repo       model.ProductRepository
	dispatcher EventDispatcher
}

func (s *productService) CreateProduct(name string, price float64) (*model.Product, error) {
	if name == "" {
		return nil, model.ErrProductNameRequired
	}
	if price < 0 {
		return nil, model.ErrProductPriceInvalid
	}

	if _, err := s.repo.FindByName(name); !errors.Is(err, model.ErrProductNotFound) {
		if err == nil {
			return nil, model.ErrProductNameExists
		}
		return nil, fmt.Errorf("failed to check product name existence: %w", err)
	}

	id, err := s.repo.NextID()
	if err != nil {
		return nil, fmt.Errorf("failed to get next product id: %w", err)
	}
	now := time.Now()
	product := &model.Product{
		ID:        id,
		Name:      name,
		Price:     price,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Store(product); err != nil {
		return nil, fmt.Errorf("failed to store product: %w", err)
	}

	event := model.ProductCreated{ProductID: product.ID, Name: product.Name, Price: product.Price}
	if err := s.dispatcher.Dispatch(event); err != nil {
		fmt.Printf("warning: failed to dispatch ProductCreated event: %v\n", err)
	}

	return product, nil
}

func (s *productService) UpdateProduct(id uuid.UUID, name string, price float64) (*model.Product, error) {
	if name == "" {
		return nil, model.ErrProductNameRequired
	}
	if price < 0 {
		return nil, model.ErrProductPriceInvalid
	}

	product, err := s.repo.Find(id)
	if err != nil {
		return nil, err
	}

	oldName := product.Name
	oldPrice := product.Price

	if existing, err := s.repo.FindByName(name); err == nil && existing.ID != id {
		return nil, model.ErrProductNameExists
	}

	product.Name = name
	product.Price = price
	product.UpdatedAt = time.Now()

	if err := s.repo.Store(product); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	event := model.ProductUpdated{
		ProductID: product.ID,
		OldName:   oldName,
		NewName:   product.Name,
		OldPrice:  oldPrice,
		NewPrice:  product.Price,
	}
	if err := s.dispatcher.Dispatch(event); err != nil {
		fmt.Printf("warning: failed to dispatch ProductUpdated event: %v\n", err)
	}

	return product, nil
}

func (s *productService) DeleteProduct(id uuid.UUID) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	event := model.ProductDeleted{ProductID: id}
	if err := s.dispatcher.Dispatch(event); err != nil {
		fmt.Printf("warning: failed to dispatch ProductDeleted event: %v\n", err)
	}
	return nil
}

func (s *productService) GetProduct(id uuid.UUID) (*model.Product, error) {
	return s.repo.Find(id)
}

func (s *productService) ListAllProducts() ([]*model.Product, error) {
	return s.repo.ListAll()
}
