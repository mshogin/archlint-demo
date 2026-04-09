//go:build ignore

// step0-behavior-clean.go - OrderService WITHOUT behavioral cycle.
//
// This is the clean version: OrderService and InventoryService
// do not call each other. Dependencies flow in one direction only.
//
// Call graph from CreateOrder:
//
//	CreateOrder -> inventoryCheck (local, no cross-service call)
//	CreateOrder -> repo.Save
//
// archlint callgraph output:
//
//	entry_point: internal/service.OrderService.CreateOrder
//	nodes: 4
//	cycles_detected: 0
//
// Use this file to show the baseline before the cycle is introduced.
package service

import (
	"demo/internal/model"
	"fmt"
)

// InventoryChecker is a narrow interface - OrderService depends on abstraction.
// This is the clean pattern: the dependency flows inward, not back out.
type InventoryChecker interface {
	CheckStock(productID string, qty int) bool
}

// OrderServiceClean is the clean version - no cycle.
// It depends on InventoryChecker (interface), not on InventoryService (concrete).
// InventoryService does NOT call back into OrderService.
type OrderServiceClean struct {
	repo      OrderRepository
	inventory InventoryChecker
}

// NewOrderServiceClean creates an OrderServiceClean.
func NewOrderServiceClean(repo OrderRepository, inv InventoryChecker) *OrderServiceClean {
	return &OrderServiceClean{repo: repo, inventory: inv}
}

// CreateOrder creates a new order, checking inventory through the interface.
// Call graph: CreateOrder -> inventory.CheckStock (leaf) -> repo.Save (leaf)
// No cycle. No mutual dependency.
func (s *OrderServiceClean) CreateOrder(userID string, items []model.OrderItem) (*model.Order, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}

	// Clean: call through interface, not through concrete InventoryService.
	// InventoryService.CheckStock does not call back into OrderService.
	for _, item := range items {
		if !s.inventory.CheckStock(item.ProductID, item.Quantity) {
			return nil, fmt.Errorf("product %s not available", item.ProductID)
		}
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

// InventoryServiceClean checks stock without calling back into OrderService.
// This is the clean version: no mutual dependency.
type InventoryServiceClean struct {
	stock map[string]int
}

// CheckStock checks whether a product has sufficient stock.
// Leaf function: no calls back into order logic.
// archlint callgraph: no outbound edges to order service.
func (s *InventoryServiceClean) CheckStock(productID string, qty int) bool {
	available, ok := s.stock[productID]
	return ok && available >= qty
}
