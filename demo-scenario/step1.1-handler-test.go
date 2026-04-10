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

// TestHandlerGetOrderCached verifies that GetOrder populates the cache.
// Uses a separate cache instance for the handler's service so we know
// the cache is populated by Get, not by Create.
// Fails if cache logic is removed entirely instead of moved to the correct layer.
func TestHandlerGetOrderCached(t *testing.T) {
	// Create the order using one service instance (its cache is irrelevant here).
	orderRepo := repo.NewOrderRepo()
	createSvc := service.NewOrderService(orderRepo, repo.NewOrderCache())
	items := []model.OrderItem{{ProductID: "p-001", Quantity: 1, Price: 15.00}}
	created, err := createSvc.Create("user-test", items)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Handler uses a fresh service with an empty cache.
	getCache := repo.NewOrderCache()
	getSvc := service.NewOrderService(orderRepo, getCache)
	h := handler.NewOrderHandler(getSvc)

	req := httptest.NewRequest(http.MethodGet, "/orders/"+created.ID, nil)
	w := httptest.NewRecorder()
	h.GetOrder(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// After GetOrder, the order must be in cache.
	// Fails if cache was removed instead of moved to the service layer.
	if _, ok := getCache.Get(created.ID); !ok {
		t.Error("GetOrder must populate the cache: cache logic missing or in wrong layer")
	}
}
