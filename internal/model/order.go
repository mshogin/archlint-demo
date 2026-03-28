package model

import "time"

// Order represents a customer order.
type Order struct {
	ID        string
	UserID    string
	Items     []OrderItem
	Total     float64
	Status    string
	CreatedAt time.Time
}

// OrderItem is a single line in an order.
type OrderItem struct {
	ProductID string
	Quantity  int
	Price     float64
}

// CreateOrderRequest is the payload for POST /orders.
type CreateOrderRequest struct {
	UserID string
	Items  []OrderItem
}
