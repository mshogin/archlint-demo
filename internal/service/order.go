package service

import (
	"demo/internal/model"
	"fmt"
	"math/rand"
	"time"
)

// OrderRepository is the narrow interface OrderService depends on.
type OrderRepository interface {
	Save(o *model.Order) error
	FindByID(id string) (*model.Order, error)
}

// OrderCache is the narrow interface for order caching.
// Service depends on this abstraction, not on a concrete repo type.
// This is how cache logic stays inside the service layer.
type OrderCache interface {
	Get(id string) (*model.Order, bool)
	Set(o *model.Order)
}

// OrderService implements order business logic.
type OrderService struct {
	repo  OrderRepository
	cache OrderCache
}

// NewOrderService creates an OrderService.
func NewOrderService(repo OrderRepository, cache OrderCache) *OrderService {
	return &OrderService{repo: repo, cache: cache}
}

// Create validates and persists a new order.
func (s *OrderService) Create(userID string, items []model.OrderItem) (*model.Order, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("items are required")
	}

	var total float64
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}

	o := &model.Order{
		ID:        fmt.Sprintf("ord-%d-%d", time.Now().UnixNano(), rand.Int63()),
		UserID:    userID,
		Items:     items,
		Total:     total,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	if err := s.repo.Save(o); err != nil {
		return nil, fmt.Errorf("saving order: %w", err)
	}

	if s.cache != nil {
		s.cache.Set(o)
	}

	return o, nil
}

// Get retrieves an order by ID, using cache when available.
func (s *OrderService) Get(id string) (*model.Order, error) {
	if s.cache != nil {
		if o, ok := s.cache.Get(id); ok {
			return o, nil
		}
	}
	o, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		s.cache.Set(o)
	}
	return o, nil
}
