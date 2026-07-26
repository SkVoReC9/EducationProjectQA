package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"awesomeProject/internal/repository"
)

var ErrPromocodeNotFound = repository.ErrPromocodeNotFound

type PromoRepository interface {
	GetPromocode(code string) (repository.Promocode, error)
	ListPromocodes() ([]repository.Promocode, error)
	CreatePromocode(p repository.Promocode) (repository.Promocode, error)
	UpdatePromocode(p repository.Promocode) (repository.Promocode, error)
	DeletePromocode(code string) error
}

type PromoService struct {
	repo PromoRepository
}

func NewPromoService(repo PromoRepository) *PromoService {
	return &PromoService{repo: repo}
}

func (s *PromoService) GetPromocode(code string) (repository.Promocode, error) {
	return s.repo.GetPromocode(code)
}

func (s *PromoService) List() ([]repository.Promocode, error) {
	return s.repo.ListPromocodes()
}

func (s *PromoService) Create(code, discountType string, value int64, active bool, expiresAt *time.Time) (repository.Promocode, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return repository.Promocode{}, errors.New("code обязателен")
	}
	if discountType != repository.DiscountPercent && discountType != repository.DiscountFixedCents {
		return repository.Promocode{}, errors.New("discount_type должен быть percent или fixed_cents")
	}
	if value <= 0 {
		return repository.Promocode{}, errors.New("discount_value должен быть > 0")
	}
	if discountType == repository.DiscountPercent && value > 100 {
		return repository.Promocode{}, errors.New("percent не может быть больше 100")
	}

	return s.repo.CreatePromocode(repository.Promocode{
		Code:          code,
		DiscountType:  discountType,
		DiscountValue: value,
		Active:        active,
		ExpiresAt:     expiresAt,
	})
}

func (s *PromoService) Update(code, discountType string, value int64, active bool, expiresAt *time.Time) (repository.Promocode, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return repository.Promocode{}, errors.New("code обязателен")
	}
	if discountType != repository.DiscountPercent && discountType != repository.DiscountFixedCents {
		return repository.Promocode{}, errors.New("discount_type должен быть percent или fixed_cents")
	}
	if value <= 0 {
		return repository.Promocode{}, errors.New("discount_value должен быть > 0")
	}
	if discountType == repository.DiscountPercent && value > 100 {
		return repository.Promocode{}, errors.New("percent не может быть больше 100")
	}

	saved, err := s.repo.UpdatePromocode(repository.Promocode{
		Code:          code,
		DiscountType:  discountType,
		DiscountValue: value,
		Active:        active,
		ExpiresAt:     expiresAt,
	})
	if err != nil {
		return repository.Promocode{}, err
	}
	return saved, nil
}

func (s *PromoService) Delete(code string) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return errors.New("code обязателен")
	}
	if err := s.repo.DeletePromocode(code); err != nil {
		return err
	}
	return nil
}

func ParseExpiresAt(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil, fmt.Errorf("expires_at должен быть RFC3339: %w", err)
	}
	return &t, nil
}
