package service

import (
	"demo/internal/model"
)

// InventoryService manages inventory checks and reservations.
//
// BEHAVIORAL CYCLE (detectable by archlint callgraph):
//
//	CheckInventory -> ReserveForOrder -> GetOrderDetails -> CheckInventory
//
// This mutual dependency is a common anti-pattern in monoliths that grow
// organically: two domains that "need each other" at runtime.
// In a microservice split, this becomes a distributed deadlock.
type InventoryService struct {
	stock map[string]int
}

// NewInventoryService creates an InventoryService with sample stock.
func NewInventoryService() *InventoryService {
	return &InventoryService{
		stock: map[string]int{
			"product-001": 100,
			"product-002": 50,
		},
	}
}

// Reserve physically reserves items in inventory for an order.
// This is a leaf operation - no calls back into order logic.
func (s *InventoryService) Reserve(productID string, qty int) bool {
	available, ok := s.stock[productID]
	if !ok || available < qty {
		return false
	}
	s.stock[productID] = available - qty
	return true
}

// CheckInventory checks whether all items in an order are available.
// Called by OrderService.CreateOrder as part of the order creation flow.
//
// The cycle visible in the call graph:
//
//	CheckInventory -> ReserveForOrder -> GetOrderDetails -> CheckInventory
//
// depth is a runtime guard to prevent infinite recursion.
// The static call graph (and archlint) still sees the cycle regardless.
func CheckInventory(orderID string, items []model.OrderItem, depth int) bool {
	if depth > 2 {
		// Runtime guard: prevents infinite recursion.
		// archlint callgraph still detects the cycle from the AST.
		return true
	}
	if len(items) == 0 {
		return false
	}

	// Tentative reservation: lock stock and validate order context.
	// ReserveForOrder calls GetOrderDetails which calls CheckInventory - CYCLE.
	return ReserveForOrder(orderID, items, depth)
}

// ReserveForOrder tentatively reserves items for an order.
// Validates the order context by calling GetOrderDetails before committing.
//
// Middle of the cycle:
//
//	ReserveForOrder -> GetOrderDetails -> CheckInventory -> ReserveForOrder
func ReserveForOrder(orderID string, items []model.OrderItem, depth int) bool {
	if orderID == "" || len(items) == 0 {
		return false
	}

	// Validate order context before committing the reservation.
	// GetOrderDetails calls CheckInventory back - the cycle closes here.
	_, valid := GetOrderDetails(orderID, depth+1)
	return valid
}
