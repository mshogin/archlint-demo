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
	if len(items) == 0 {
		return fmt.Errorf("no items to fulfill")
	}
	if orderID == "" {
		return fmt.Errorf("orderID required")
	}
	return nil
}
