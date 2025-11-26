package tests

import (
	"testing"
	"time"

	"product/pkg/domain/model"
	"product/pkg/domain/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ model.ProductRepository = (*mockProductRepository)(nil)

type mockProductRepository struct {
	storeByID   map[uuid.UUID]*model.Product
	storeByName map[string]*model.Product
}

func newMockProductRepository() *mockProductRepository {
	return &mockProductRepository{
		storeByID:   make(map[uuid.UUID]*model.Product),
		storeByName: make(map[string]*model.Product),
	}
}

func (m *mockProductRepository) NextID() (uuid.UUID, error) { return uuid.NewV7() }
func (m *mockProductRepository) Store(p *model.Product) error {
	pCopy := *p
	for name, prod := range m.storeByName {
		if prod.ID == p.ID && name != p.Name {
			delete(m.storeByName, name)
		}
	}
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

type mockEventDispatcher struct{ events []service.Event }

func (m *mockEventDispatcher) Dispatch(event service.Event) error {
	m.events = append(m.events, event)
	return nil
}
func (m *mockEventDispatcher) Clear() { m.events = nil }

func TestProductService_CreateProduct(t *testing.T) {
	t.Run("successfully creates a product", func(t *testing.T) {
		repo := newMockProductRepository()
		dispatcher := &mockEventDispatcher{}
		productService := service.NewProductService(repo, dispatcher)

		name, price := "Test Laptop", 1200.50
		product, err := productService.CreateProduct(name, price)

		require.NoError(t, err)
		assert.Equal(t, name, product.Name)
		assert.Equal(t, price, product.Price)

		stored, _ := repo.Find(product.ID)
		assert.NotNil(t, stored)

		require.Len(t, dispatcher.events, 1)
		event, ok := dispatcher.events[0].(model.ProductCreated)
		require.True(t, ok)
		assert.Equal(t, product.ID, event.ProductID)
	})

	t.Run("fails on invalid data", func(t *testing.T) {
		productService := service.NewProductService(newMockProductRepository(), &mockEventDispatcher{})

		_, err := productService.CreateProduct("", 100)
		assert.ErrorIs(t, err, model.ErrProductNameRequired)

		_, err = productService.CreateProduct("Mouse", -10)
		assert.ErrorIs(t, err, model.ErrProductPriceInvalid)
	})

	t.Run("fails on duplicate name", func(t *testing.T) {
		repo := newMockProductRepository()
		productService := service.NewProductService(repo, &mockEventDispatcher{})

		_, _ = productService.CreateProduct("Duplicate Product", 100)
		_, err := productService.CreateProduct("Duplicate Product", 200)

		assert.ErrorIs(t, err, model.ErrProductNameExists)
	})
}

func TestProductService_UpdateProduct(t *testing.T) {
	repo := newMockProductRepository()
	dispatcher := &mockEventDispatcher{}
	productService := service.NewProductService(repo, dispatcher)

	p1, _ := productService.CreateProduct("Product 1", 10.0)
	_, _ = productService.CreateProduct("Existing Name", 99.0)
	dispatcher.Clear()

	t.Run("successfully updates a product and dispatches event", func(t *testing.T) {
		newName, newPrice := "New Name for P1", 15.5
		updated, err := productService.UpdateProduct(p1.ID, newName, newPrice)

		require.NoError(t, err)
		assert.Equal(t, newName, updated.Name)
		assert.Equal(t, newPrice, updated.Price)

		stored, _ := repo.Find(p1.ID)
		assert.Equal(t, newName, stored.Name)

		require.Len(t, dispatcher.events, 1)
		event, ok := dispatcher.events[0].(model.ProductUpdated)
		require.True(t, ok)
		assert.Equal(t, p1.ID, event.ProductID)
		assert.Equal(t, "Product 1", event.OldName)
		assert.Equal(t, newName, event.NewName)
		assert.Equal(t, 10.0, event.OldPrice)
		assert.Equal(t, newPrice, event.NewPrice)
	})

	t.Run("fails to update non-existent product", func(t *testing.T) {
		dispatcher.Clear()
		_, err := productService.UpdateProduct(uuid.New(), "any name", 10)
		assert.ErrorIs(t, err, model.ErrProductNotFound)
		assert.Empty(t, dispatcher.events)
	})

	t.Run("fails on invalid data", func(t *testing.T) {
		dispatcher.Clear()
		_, err := productService.UpdateProduct(p1.ID, "", 20)
		assert.ErrorIs(t, err, model.ErrProductNameRequired)
		assert.Empty(t, dispatcher.events)
	})

	t.Run("fails on duplicate name with another product", func(t *testing.T) {
		dispatcher.Clear()
		_, err := productService.UpdateProduct(p1.ID, "Existing Name", 20)
		assert.ErrorIs(t, err, model.ErrProductNameExists)
		assert.Empty(t, dispatcher.events)
	})
}

func TestProductService_DeleteProduct(t *testing.T) {
	t.Run("successfully deletes a product", func(t *testing.T) {
		repo := newMockProductRepository()
		dispatcher := &mockEventDispatcher{}
		productService := service.NewProductService(repo, dispatcher)
		product, _ := productService.CreateProduct("To Be Deleted", 50)
		dispatcher.Clear()

		err := productService.DeleteProduct(product.ID)
		require.NoError(t, err)

		_, err = repo.Find(product.ID)
		assert.ErrorIs(t, err, model.ErrProductNotFound)

		require.Len(t, dispatcher.events, 1)
		event, ok := dispatcher.events[0].(model.ProductDeleted)
		require.True(t, ok)
		assert.Equal(t, product.ID, event.ProductID)
	})

	t.Run("fails to delete non-existent product", func(t *testing.T) {
		repo := newMockProductRepository()
		dispatcher := &mockEventDispatcher{}
		productService := service.NewProductService(repo, dispatcher)

		err := productService.DeleteProduct(uuid.New())
		assert.ErrorIs(t, err, model.ErrProductNotFound)
		assert.Empty(t, dispatcher.events)
	})
}

func TestProductService_ListAllProducts(t *testing.T) {
	repo := newMockProductRepository()
	productService := service.NewProductService(repo, &mockEventDispatcher{})

	p1, _ := productService.CreateProduct("Product 1", 10)
	p2, _ := productService.CreateProduct("Product 2", 20)
	p3, _ := productService.CreateProduct("Product 3 (deleted)", 30)
	_ = productService.DeleteProduct(p3.ID)

	products, err := productService.ListAllProducts()
	require.NoError(t, err)
	assert.Len(t, products, 2)

	productIDs := make([]uuid.UUID, len(products))
	for i, p := range products {
		productIDs[i] = p.ID
	}
	assert.ElementsMatch(t, []uuid.UUID{p1.ID, p2.ID}, productIDs)
}
