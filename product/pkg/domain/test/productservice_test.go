package tests

import (
	"testing"
	"time"

	"product/pkg/domain/model"
	"product/pkg/domain/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestProductService(t *testing.T) {
	t.Run("CreateProduct_SuccessfullyCreatesAProduct", func(t *testing.T) {
		repo := &mockProductRepository{
			storeByID:   make(map[uuid.UUID]*model.Product),
			storeByName: make(map[string]*model.Product),
		}
		dispatcher := &mockEventDispatcher{}
		svc := service.NewProductService(repo, dispatcher)

		name, price := "Test Laptop", 1200.50
		product, err := svc.CreateProduct(name, price)
		require.NoError(t, err)
		require.Equal(t, name, product.Name)
		require.Equal(t, price, product.Price)

		stored, err := repo.Find(product.ID)
		require.NoError(t, err)
		require.NotNil(t, stored)

		require.Len(t, dispatcher.events, 1)
		event, ok := dispatcher.events[0].(model.ProductCreated)
		require.True(t, ok)
		require.Equal(t, product.ID, event.ProductID)
	})

	t.Run("CreateProduct_FailsOnEmptyName", func(t *testing.T) {
		repo := &mockProductRepository{
			storeByID:   make(map[uuid.UUID]*model.Product),
			storeByName: make(map[string]*model.Product),
		}
		dispatcher := &mockEventDispatcher{}
		svc := service.NewProductService(repo, dispatcher)

		_, err := svc.CreateProduct("", 100)
		require.ErrorIs(t, err, model.ErrProductNameRequired)
	})

	t.Run("CreateProduct_FailsOnNegativePrice", func(t *testing.T) {
		repo := &mockProductRepository{
			storeByID:   make(map[uuid.UUID]*model.Product),
			storeByName: make(map[string]*model.Product),
		}
		dispatcher := &mockEventDispatcher{}
		svc := service.NewProductService(repo, dispatcher)

		_, err := svc.CreateProduct("Mouse", -10)
		require.ErrorIs(t, err, model.ErrProductPriceInvalid)
	})

	t.Run("CreateProduct_FailsOnDuplicateName", func(t *testing.T) {
		repo := &mockProductRepository{
			storeByID:   make(map[uuid.UUID]*model.Product),
			storeByName: make(map[string]*model.Product),
		}
		dispatcher := &mockEventDispatcher{}
		svc := service.NewProductService(repo, dispatcher)

		_, _ = svc.CreateProduct("Duplicate Product", 100)
		_, err := svc.CreateProduct("Duplicate Product", 200)
		require.ErrorIs(t, err, model.ErrProductNameExists)
	})

	t.Run("UpdateProduct_SuccessfullyUpdatesAProduct", func(t *testing.T) {
		repo := &mockProductRepository{
			storeByID:   make(map[uuid.UUID]*model.Product),
			storeByName: make(map[string]*model.Product),
		}
		dispatcher := &mockEventDispatcher{}
		svc := service.NewProductService(repo, dispatcher)

		product, err := svc.CreateProduct("Product 1", 10.0)
		require.NoError(t, err)

		newName, newPrice := "New Name for P1", 15.5
		updated, err := svc.UpdateProduct(product.ID, newName, newPrice)
		require.NoError(t, err)
		require.Equal(t, newName, updated.Name)
		require.Equal(t, newPrice, updated.Price)

		stored, err := repo.Find(product.ID)
		require.NoError(t, err)
		require.Equal(t, newName, stored.Name)

		require.Len(t, dispatcher.events, 2)
		event, ok := dispatcher.events[1].(model.ProductUpdated)
		require.True(t, ok)
		require.Equal(t, product.ID, event.ProductID)
		require.Equal(t, "Product 1", event.OldName)
		require.Equal(t, newName, event.NewName)
		require.Equal(t, 10.0, event.OldPrice)
		require.Equal(t, newPrice, event.NewPrice)
	})

	t.Run("UpdateProduct_FailsOnNonExistentProduct", func(t *testing.T) {
		repo := &mockProductRepository{
			storeByID:   make(map[uuid.UUID]*model.Product),
			storeByName: make(map[string]*model.Product),
		}
		dispatcher := &mockEventDispatcher{}
		svc := service.NewProductService(repo, dispatcher)

		_, err := svc.UpdateProduct(uuid.New(), "any name", 10)
		require.ErrorIs(t, err, model.ErrProductNotFound)
		require.Empty(t, dispatcher.events)
	})

	t.Run("UpdateProduct_FailsOnEmptyName", func(t *testing.T) {
		repo := &mockProductRepository{
			storeByID:   make(map[uuid.UUID]*model.Product),
			storeByName: make(map[string]*model.Product),
		}
		dispatcher := &mockEventDispatcher{}
		svc := service.NewProductService(repo, dispatcher)

		product, err := svc.CreateProduct("Valid", 50)
		require.NoError(t, err)
		dispatcher.Clear()

		_, err = svc.UpdateProduct(product.ID, "", 20)
		require.ErrorIs(t, err, model.ErrProductNameRequired)
		require.Empty(t, dispatcher.events)
	})

	t.Run("UpdateProduct_FailsOnDuplicateName", func(t *testing.T) {
		repo := &mockProductRepository{
			storeByID:   make(map[uuid.UUID]*model.Product),
			storeByName: make(map[string]*model.Product),
		}
		dispatcher := &mockEventDispatcher{}
		svc := service.NewProductService(repo, dispatcher)

		p1, err := svc.CreateProduct("Product A", 10)
		require.NoError(t, err)
		_, err = svc.CreateProduct("Existing Name", 20)
		require.NoError(t, err)
		dispatcher.Clear()

		_, err = svc.UpdateProduct(p1.ID, "Existing Name", 30)
		require.ErrorIs(t, err, model.ErrProductNameExists)
		require.Empty(t, dispatcher.events)
	})

	t.Run("DeleteProduct_SuccessfullyDeletesAProduct", func(t *testing.T) {
		repo := &mockProductRepository{
			storeByID:   make(map[uuid.UUID]*model.Product),
			storeByName: make(map[string]*model.Product),
		}
		dispatcher := &mockEventDispatcher{}
		svc := service.NewProductService(repo, dispatcher)

		product, err := svc.CreateProduct("To Be Deleted", 50)
		require.NoError(t, err)
		dispatcher.Clear()

		err = svc.DeleteProduct(product.ID)
		require.NoError(t, err)

		_, err = repo.Find(product.ID)
		require.ErrorIs(t, err, model.ErrProductNotFound)

		require.Len(t, dispatcher.events, 1)
		event, ok := dispatcher.events[0].(model.ProductDeleted)
		require.True(t, ok)
		require.Equal(t, product.ID, event.ProductID)
	})

	t.Run("DeleteProduct_FailsOnNonExistentProduct", func(t *testing.T) {
		repo := &mockProductRepository{
			storeByID:   make(map[uuid.UUID]*model.Product),
			storeByName: make(map[string]*model.Product),
		}
		dispatcher := &mockEventDispatcher{}
		svc := service.NewProductService(repo, dispatcher)

		err := svc.DeleteProduct(uuid.New())
		require.ErrorIs(t, err, model.ErrProductNotFound)
		require.Empty(t, dispatcher.events)
	})

	t.Run("ListAllProducts_ReturnsOnlyActiveProducts", func(t *testing.T) {
		repo := &mockProductRepository{
			storeByID:   make(map[uuid.UUID]*model.Product),
			storeByName: make(map[string]*model.Product),
		}
		dispatcher := &mockEventDispatcher{}
		svc := service.NewProductService(repo, dispatcher)

		p1, err := svc.CreateProduct("Product 1", 10)
		require.NoError(t, err)
		p2, err := svc.CreateProduct("Product 2", 20)
		require.NoError(t, err)
		p3, err := svc.CreateProduct("Product 3 (deleted)", 30)
		require.NoError(t, err)
		err = svc.DeleteProduct(p3.ID)
		require.NoError(t, err)

		products, err := svc.ListAllProducts()
		require.NoError(t, err)
		require.Len(t, products, 2)

		var foundIDs []uuid.UUID
		for _, p := range products {
			foundIDs = append(foundIDs, p.ID)
		}
		require.ElementsMatch(t, []uuid.UUID{p1.ID, p2.ID}, foundIDs)
	})
}

var _ model.ProductRepository = (*mockProductRepository)(nil)

type mockProductRepository struct {
	storeByID   map[uuid.UUID]*model.Product
	storeByName map[string]*model.Product
}

func (m *mockProductRepository) NextID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (m *mockProductRepository) Store(p *model.Product) error {
	now := time.Now()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now

	for name, existing := range m.storeByName {
		if existing.ID == p.ID && name != p.Name {
			delete(m.storeByName, name)
			break
		}
	}

	pCopy := *p
	m.storeByID[p.ID] = &pCopy
	m.storeByName[p.Name] = &pCopy
	return nil
}

func (m *mockProductRepository) Find(id uuid.UUID) (*model.Product, error) {
	p, ok := m.storeByID[id]
	if !ok || p.DeletedAt != nil {
		return nil, model.ErrProductNotFound
	}
	return p, nil
}

func (m *mockProductRepository) FindByName(name string) (*model.Product, error) {
	p, ok := m.storeByName[name]
	if !ok || p.DeletedAt != nil {
		return nil, model.ErrProductNotFound
	}
	return p, nil
}

func (m *mockProductRepository) Delete(id uuid.UUID) error {
	p, ok := m.storeByID[id]
	if !ok || p.DeletedAt != nil {
		return model.ErrProductNotFound
	}
	now := time.Now()
	p.DeletedAt = &now
	p.UpdatedAt = now
	delete(m.storeByName, p.Name)
	return m.Store(p)
}

func (m *mockProductRepository) ListAll() ([]*model.Product, error) {
	var products []*model.Product
	for _, p := range m.storeByID {
		if p.DeletedAt == nil {
			products = append(products, p)
		}
	}
	return products, nil
}

var _ service.EventDispatcher = (*mockEventDispatcher)(nil)

type mockEventDispatcher struct {
	events []service.Event
}

func (m *mockEventDispatcher) Dispatch(e service.Event) error {
	m.events = append(m.events, e)
	return nil
}

func (m *mockEventDispatcher) Clear() {
	m.events = nil
}
