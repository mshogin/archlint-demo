//go:build ignore

// step1-behavior-cycle.go - OrderService WITH behavioral cycle.
//
// The cycle is introduced when inventory logic starts calling back
// into order logic "just to validate context". This is the most
// common way behavioral cycles appear in practice.
//
// Call graph from CreateOrder:
//
//	CreateOrder
//	  -> CheckInventory
//	      -> ReserveForOrder
//	          -> GetOrderDetails     <- crosses back into order domain
//	              -> CheckInventory  <- CYCLE (archlint reports cycle: true)
//
// archlint callgraph output:
//
//	entry_point: internal/service.OrderService.CreateOrder
//	nodes: 6
//	cycles_detected: 1
//	  GetOrderDetails -> CheckInventory  [cycle: true]
//
// To detect with archlint:
//
//	archlint callgraph ./internal --entry "internal/service.OrderService.CreateOrder" --no-puml
//
// The cycle means:
//   - You cannot split OrderService and InventoryService into separate services
//     without breaking the call chain.
//   - Changing CheckInventory may break GetOrderDetails behavior.
//   - Changing GetOrderDetails may change what CheckInventory returns.
//   - Integration tests must run both services together - they are coupled at runtime.
package service

import (
	"demo/internal/model"
	"fmt"
)

// OrderServiceWithCycle demonstrates the anti-pattern.
// CreateOrder calls CheckInventory, which eventually calls GetOrderDetails,
// which calls CheckInventory again - mutual dependency between order and inventory domains.
type OrderServiceWithCycle struct {
	repo OrderRepository
}

// NewOrderServiceWithCycle creates an OrderServiceWithCycle.
func NewOrderServiceWithCycle(repo OrderRepository) *OrderServiceWithCycle {
	return &OrderServiceWithCycle{repo: repo}
}

// CreateOrderWithCycle is the entry point for the behavioral cycle demo.
//
// archlint callgraph --entry "internal/service.OrderServiceWithCycle.CreateOrderWithCycle"
// will report cycles_detected: 1
func (s *OrderServiceWithCycle) CreateOrderWithCycle(userID string, items []model.OrderItem) (*model.Order, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}

	// This single call initiates a chain that cycles back to itself.
	if !checkInventoryCycle(userID, items, 0) {
		return nil, fmt.Errorf("items not available in inventory")
	}

	var total float64
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}

	o := &model.Order{
		ID:     fmt.Sprintf("ord-%s-001", userID),
		UserID: userID,
		Items:  items,
		Total:  total,
		Status: "pending",
	}

	return o, s.repo.Save(o)
}

// getOrderDetailsCycle retrieves order context.
// PROBLEM: calls back into checkInventoryCycle - creates the cycle.
// Once a developer adds this "just to validate", the cycle is born.
func getOrderDetailsCycle(orderID string, depth int) (*model.Order, bool) {
	// This "validation" call back into inventory creates the mutual dependency.
	items := []model.OrderItem{{ProductID: "product-001", Quantity: 1, Price: 10.0}}
	available := checkInventoryCycle(orderID, items, depth+1) // <- CYCLE CLOSES HERE
	return &model.Order{ID: orderID, Status: "active"}, available
}

// checkInventoryCycle checks stock and validates via ReserveForOrder.
// Part of the cycle: checkInventoryCycle -> reserveForOrderCycle -> getOrderDetailsCycle -> checkInventoryCycle
func checkInventoryCycle(orderID string, items []model.OrderItem, depth int) bool {
	if depth > 2 {
		return true // depth guard prevents runtime infinite recursion
	}
	if len(items) == 0 {
		return false
	}
	return reserveForOrderCycle(orderID, items, depth)
}

// reserveForOrderCycle tentatively reserves items, calling GetOrderDetails to validate.
// This is where inventory "reaches into" order domain - the design smell.
func reserveForOrderCycle(orderID string, items []model.OrderItem, depth int) bool {
	if orderID == "" || len(items) == 0 {
		return false
	}
	// "I need to check the order is still valid before reserving."
	// Sounds reasonable. Creates the cycle.
	_, valid := getOrderDetailsCycle(orderID, depth+1) // <- crosses into order domain
	return valid
}

// CheckInventory is the exported entry point called by OrderService.CreateOrder.
// Cycle version: delegates to checkInventoryCycle which eventually calls GetOrderDetails
// which calls CheckInventory again - creating the behavioral cycle.
func CheckInventory(userID string, items []model.OrderItem, depth int) bool {
	return checkInventoryCycle(userID, items, depth)
}
