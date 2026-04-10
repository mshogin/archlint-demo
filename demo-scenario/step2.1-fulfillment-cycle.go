//go:build ignore

package service

import (
	"demo/internal/model"
	"fmt"
)

// FulfillmentService manages the order fulfillment flow.
type FulfillmentService struct{}

func NewFulfillmentService() *FulfillmentService {
	return &FulfillmentService{}
}

// FulfillOrder validates items and fulfills the order.
func (f *FulfillmentService) FulfillOrder(orderID string, items []model.OrderItem) error {
	return fulfillOrder(orderID, items, 0)
}

func fulfillOrder(orderID string, items []model.OrderItem, depth int) error {
	if depth > 2 {
		return nil
	}
	if len(items) == 0 {
		return fmt.Errorf("no items to fulfill")
	}
	return reserveForFulfillment(orderID, items, depth)
}

func reserveForFulfillment(orderID string, items []model.OrderItem, depth int) error {
	if orderID == "" {
		return fmt.Errorf("orderID required")
	}
	return validateFulfillmentContext(orderID, items, depth)
}

func validateFulfillmentContext(orderID string, items []model.OrderItem, depth int) error {
	return fulfillOrder(orderID, items, depth+1)
}
