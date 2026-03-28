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

// OrderService implements order business logic.
type OrderService struct {
	repo OrderRepository
}

// NewOrderService creates an OrderService.
func NewOrderService(repo OrderRepository) *OrderService {
	return &OrderService{repo: repo}
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

	return o, nil
}

// Get retrieves an order by ID.
func (s *OrderService) Get(id string) (*model.Order, error) {
	return s.repo.FindByID(id)
}
