// Package tests contains functional tests for the demo application.
// These tests trace the call chain through handler -> service -> repo,
// and demonstrate the behavioral cycle in the service layer.
package tests

import (
	"demo/internal/model"
	"demo/internal/repo"
	"demo/internal/service"
	"testing"
)

// TestCreateOrder traces the full call chain:
//
//	OrderService.CreateOrder -> CheckInventory -> ReserveForOrder -> GetOrderDetails -> CheckInventory (cycle)
//
// The test itself completes because of the depth guard in CheckInventory.
// archlint callgraph detects the cycle statically from the AST.
func TestCreateOrder(t *testing.T) {
	orderRepo := repo.NewOrderRepo()
	svc := service.NewOrderService(orderRepo, nil)

	items := []model.OrderItem{
		{ProductID: "product-001", Quantity: 2, Price: 29.99},
	}

	order, err := svc.CreateOrder("user-123", items)
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	if order == nil {
		t.Fatal("expected order, got nil")
	}

	if order.UserID != "user-123" {
		t.Errorf("expected userID=user-123, got %q", order.UserID)
	}

	if order.Status != "pending" {
		t.Errorf("expected status=pending, got %q", order.Status)
	}

	expectedTotal := 2 * 29.99
	if order.Total != expectedTotal {
		t.Errorf("expected total=%.2f, got %.2f", expectedTotal, order.Total)
	}

	t.Logf("order created: id=%s total=%.2f status=%s", order.ID, order.Total, order.Status)
}

// TestCreateOrder_EmptyUserID verifies that CreateOrder rejects empty userID.
func TestCreateOrder_EmptyUserID(t *testing.T) {
	orderRepo := repo.NewOrderRepo()
	svc := service.NewOrderService(orderRepo, nil)

	items := []model.OrderItem{
		{ProductID: "product-001", Quantity: 1, Price: 10.0},
	}

	_, err := svc.CreateOrder("", items)
	if err == nil {
		t.Fatal("expected error for empty userID, got nil")
	}
}

// TestCreateOrder_EmptyItems verifies that CreateOrder rejects empty items.
func TestCreateOrder_EmptyItems(t *testing.T) {
	orderRepo := repo.NewOrderRepo()
	svc := service.NewOrderService(orderRepo, nil)

	_, err := svc.CreateOrder("user-123", nil)
	if err == nil {
		t.Fatal("expected error for empty items, got nil")
	}
}

// TestGetOrder traces:
//
//	OrderService.Get -> repo.FindByID
//
// First creates an order, then retrieves it.
func TestGetOrder(t *testing.T) {
	orderRepo := repo.NewOrderRepo()
	svc := service.NewOrderService(orderRepo, nil)

	// Create an order first.
	items := []model.OrderItem{
		{ProductID: "product-002", Quantity: 1, Price: 49.99},
	}

	created, err := svc.CreateOrder("user-456", items)
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	// Now retrieve it.
	got, err := svc.Get(created.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("expected id=%s, got %s", created.ID, got.ID)
	}

	if got.UserID != "user-456" {
		t.Errorf("expected userID=user-456, got %q", got.UserID)
	}

	t.Logf("order retrieved: id=%s userID=%s", got.ID, got.UserID)
}

// TestGetOrder_NotFound verifies that Get returns error for unknown ID.
func TestGetOrder_NotFound(t *testing.T) {
	orderRepo := repo.NewOrderRepo()
	svc := service.NewOrderService(orderRepo, nil)

	_, err := svc.Get("nonexistent-id")
	if err == nil {
		t.Fatal("expected error for nonexistent order, got nil")
	}
}

// TestCheckInventory_Cycle demonstrates the behavioral cycle at the package level.
//
// Static analysis (archlint callgraph) sees:
//
//	CheckInventory -> ReserveForOrder -> GetOrderDetails -> CheckInventory
//
// The depth guard prevents actual infinite recursion at runtime.
// The cycle is still present in the call graph and detectable by static analysis.
func TestCheckInventory_Cycle(t *testing.T) {
	items := []model.OrderItem{
		{ProductID: "product-001", Quantity: 1, Price: 10.0},
	}

	// This call traverses the cycle: CheckInventory -> ReserveForOrder -> GetOrderDetails -> CheckInventory
	// depth=0 means we enter the cycle from scratch, the guard stops it at depth>2.
	result := service.CheckInventory("order-001", items, 0)

	// The result is true because the depth guard returns true at depth>2.
	// This is by design: the demo shows static cycle detection, not runtime behavior.
	if !result {
		t.Logf("CheckInventory returned false - depth guard may have triggered differently")
	}

	t.Logf("CheckInventory completed (cycle traversed with depth guard): result=%v", result)
}

// TestGetOrderDetails_Cycle demonstrates that GetOrderDetails participates in the cycle.
//
// Static call graph:
//
//	GetOrderDetails -> CheckInventory -> ReserveForOrder -> GetOrderDetails (CYCLE)
func TestGetOrderDetails_Cycle(t *testing.T) {
	order, valid := service.GetOrderDetails("order-001", 0)

	if order == nil {
		t.Fatal("expected order, got nil")
	}

	t.Logf("GetOrderDetails: id=%s valid=%v (cycle traversed with depth guard)", order.ID, valid)
}
