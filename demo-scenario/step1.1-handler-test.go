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

// TestHandlerGetOrderCached verifies that GetOrder returns a cached order correctly.
func TestHandlerGetOrderCached(t *testing.T) {
	orderRepo := repo.NewOrderRepo()
	cache := repo.NewOrderCache()
	svc := service.NewOrderService(orderRepo, cache)
	h := handler.NewOrderHandler(svc)

	items := []model.OrderItem{
		{ProductID: "p-001", Quantity: 1, Price: 15.00},
	}
	created, err := svc.Create("user-test", items)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/orders/"+created.ID, nil)
	w := httptest.NewRecorder()
	h.GetOrder(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
