// handler_cache_test.go tests the GetOrder handler with caching.
// Added in step 1.1 together with the caching feature.
package tests

import (
	"demo/internal/handler"
	"demo/internal/model"
	"demo/internal/repo"
	"demo/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHandlerGetOrderFromCache verifies that GetOrder returns an order correctly.
// The handler uses an OrderCache directly (step 1.1 pattern).
func TestHandlerGetOrderFromCache(t *testing.T) {
	orderRepo := repo.NewOrderRepo()
	cache := repo.NewOrderCache()
	svc := service.NewOrderService(orderRepo, cache)
	h := handler.NewOrderHandler(svc, cache)

	// Create an order via service so it exists in repo and cache.
	items := []model.OrderItem{
		{ProductID: "p-001", Quantity: 1, Price: 15.00},
	}
	created, err := svc.Create("user-test", items)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// GET /orders/{id} — handler should return 200 with the order.
	req := httptest.NewRequest(http.MethodGet, "/orders/"+created.ID, nil)
	w := httptest.NewRecorder()
	h.GetOrder(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}
