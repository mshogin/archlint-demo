package service

import (
	"demo/internal/model"
	"fmt"
)

// CreateOrder validates and creates a new order.
// Calls CheckInventory to verify stock availability before persisting.
//
// Entry point for the behavioral cycle demo:
//
//	CreateOrder -> CheckInventory -> ReserveForOrder -> GetOrderDetails -> CheckInventory (CYCLE)
//
// archlint callgraph detects this cycle even though it does not manifest at runtime
// as an infinite loop (the depth guard prevents that).
// In a real distributed system this becomes a deadlock: two services calling each other.
func (s *OrderService) CreateOrder(userID string, items []model.OrderItem) (*model.Order, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("items are required")
	}

	// Cross-service behavioral dependency: order creation triggers inventory check.
	// This is where the callgraph cycle begins.
	if !CheckInventory(userID, items, 0) {
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

// GetOrderDetails retrieves full order context for a given order ID.
// Called by InventoryService.ReserveForOrder to validate the order before committing stock.
//
// This function is part of the behavioral cycle:
//
//	GetOrderDetails -> CheckInventory -> ReserveForOrder -> GetOrderDetails (CYCLE)
//
// A static analyzer sees: GetOrderDetails calls CheckInventory.
// CheckInventory calls ReserveForOrder. ReserveForOrder calls GetOrderDetails again.
// This is the mutual dependency that archlint callgraph will flag.
func GetOrderDetails(orderID string, depth int) (*model.Order, bool) {
	if orderID == "" {
		return nil, false
	}

	// Validate inventory status as part of fetching order details.
	// This call back into CheckInventory closes the cycle in the call graph.
	items := []model.OrderItem{{ProductID: "product-001", Quantity: 1, Price: 10.0}}
	available := CheckInventory(orderID, items, depth+1)

	return &model.Order{ID: orderID, Status: "active"}, available
}
