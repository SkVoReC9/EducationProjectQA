package service

import "errors"

import (
	"awesomeProject/internal/repository"
)

// Интерфейс, который должен реализовать наш репозиторий
type CatalogRepository interface {
	GetProduct(id string) (repository.Product, error)
	ListProducts() ([]repository.Product, error)
}

type CatalogService struct {
	repo CatalogRepository
}

func NewCatalogService(repo CatalogRepository) *CatalogService {
	return &CatalogService{repo: repo}
}

// GetProduct возвращает товар по ID
func (s *CatalogService) GetProduct(id string) (repository.Product, error) {
	// Базовая бизнес-проверка (для ручного тестирования)
	if id == "" {
		return repository.Product{}, errors.New("product_id cannot be empty")
	}

	return s.repo.GetProduct(id)
}

// ListProducts возвращает список товаров
// В будущем сюда можно добавить логику пагинации и искусственные задержки
func (s *CatalogService) ListProducts() ([]repository.Product, error) {
	return s.repo.ListProducts()
}
