// fulfillment_test.go tests the FulfillmentService.
// Added in step 2.1 together with the fulfillment feature.
package tests

import (
	"demo/internal/model"
	"demo/internal/service"
	"testing"
)

// TestFulfillOrder verifies that FulfillOrder completes without error.
// The depth guard prevents infinite recursion even in the cycle version.
func TestFulfillOrder(t *testing.T) {
	svc := service.NewFulfillmentService()
	items := []model.OrderItem{
		{ProductID: "p-001", Quantity: 2, Price: 25.00},
	}

	if err := svc.FulfillOrder("ord-123", items); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestFulfillOrderEmptyItems(t *testing.T) {
	svc := service.NewFulfillmentService()

	err := svc.FulfillOrder("ord-123", nil)
	if err == nil {
		t.Error("expected error for empty items, got nil")
	}
}
