package memory

import (
	"errors"
	"sync"

	"awesomeProject/internal/repository"
)

var ErrNotFound = errors.New("product not found")

// CatalogRepository имитирует базу данных для каталога товаров
type CatalogRepository struct {
	mu       sync.RWMutex
	products map[string]repository.Product
}

func NewCatalogRepository() *CatalogRepository {
	return &CatalogRepository{
		products: map[string]repository.Product{
			"prod-1": {
				ID:            "prod-1",
				Name:          "iPhone 15",
				Description:   "Latest Apple smartphone",
				PriceCents:    99900,
				StockQuantity: 10,
			},
			"prod-2": {
				ID:            "prod-2",
				Name:          "MacBook Pro",
				Description:   "Apple laptop for professionals",
				PriceCents:    199900,
				StockQuantity: 5,
			},
		},
	}
}

func (r *CatalogRepository) GetProduct(id string) (repository.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	product, exists := r.products[id]
	if !exists {
		return repository.Product{}, ErrNotFound
	}
	return product, nil
}

func (r *CatalogRepository) ListProducts() ([]repository.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	products := make([]repository.Product, 0, len(r.products))
	for _, p := range r.products {
		products = append(products, p)
	}
	return products, nil
}
