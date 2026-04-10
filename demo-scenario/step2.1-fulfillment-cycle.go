//go:build ignore

package service

import (
	"demo/internal/model"
	"fmt"
)

// FulfillmentService manages the order fulfillment flow.
// Entry point for behavioral cycle demo:
//
//	FulfillOrder -> fulfillOrder -> reserveForFulfillment -> validateFulfillmentContext -> fulfillOrder (CYCLE)
//
// archlint callgraph detects:
//
//	entry_point: internal/service.FulfillmentService.FulfillOrder
//	cycles_detected: 1
//	  validateFulfillmentContext -> fulfillOrder [cycle: true]
type FulfillmentService struct{}

func NewFulfillmentService() *FulfillmentService {
	return &FulfillmentService{}
}

// FulfillOrder is the public entry point. Delegates to package-level fulfillOrder.
func (f *FulfillmentService) FulfillOrder(orderID string, items []model.OrderItem) error {
	return fulfillOrder(orderID, items, 0)
}

// fulfillOrder validates and initiates stock reservation.
func fulfillOrder(orderID string, items []model.OrderItem, depth int) error {
	if depth > 2 {
		return nil
	}
	if len(items) == 0 {
		return fmt.Errorf("no items to fulfill")
	}
	return reserveForFulfillment(orderID, items, depth)
}

// reserveForFulfillment reserves inventory for the order items.
func reserveForFulfillment(orderID string, items []model.OrderItem, depth int) error {
	if orderID == "" {
		return fmt.Errorf("orderID required")
	}
	// "Validate order state before committing the reservation"
	return validateFulfillmentContext(orderID, items, depth)
}

// validateFulfillmentContext verifies the order is still in a fulfillable state.
// PROBLEM: calls back into fulfillOrder - creates the behavioral cycle.
// A developer added this "safety check" but didn't notice it closes the loop.
func validateFulfillmentContext(orderID string, items []model.OrderItem, depth int) error {
	return fulfillOrder(orderID, items, depth+1) // <- CYCLE CLOSES HERE
}
