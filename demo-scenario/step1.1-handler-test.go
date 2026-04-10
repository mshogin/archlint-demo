// handler_cache_test.go tests that GetOrder uses caching:
// two requests for the same order must produce at most one database call.
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

// countingRepo wraps OrderRepo and counts FindByID calls.
type countingRepo struct {
	*repo.OrderRepo
	calls int
}

func (r *countingRepo) FindByID(id string) (*model.Order, error) {
	r.calls++
	return r.OrderRepo.FindByID(id)
}

// TestHandlerGetOrderCached verifies that two GETs for the same order
// produce at most one database call — the second must be served from cache.
// Fails if caching is missing or not in the correct layer.
func TestHandlerGetOrderCached(t *testing.T) {
	cr := &countingRepo{OrderRepo: repo.NewOrderRepo()}

	// Create order via a separate service (its cache is irrelevant here).
	createSvc := service.NewOrderService(cr, repo.NewOrderCache())
	items := []model.OrderItem{{ProductID: "p-001", Quantity: 1, Price: 15.00}}
	created, err := createSvc.Create("user-test", items)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	cr.calls = 0 // reset counter after create

	// Handler uses a fresh service with an empty cache.
	getSvc := service.NewOrderService(cr, repo.NewOrderCache())
	h := handler.NewOrderHandler(getSvc)

	// First GET — hits the database.
	req1 := httptest.NewRequest(http.MethodGet, "/orders/"+created.ID, nil)
	w1 := httptest.NewRecorder()
	h.GetOrder(w1, req1)
	if w1.Code != http.StatusOK {
		t.Errorf("first GET: expected 200, got %d", w1.Code)
	}

	// Second GET — must be served from cache, not the database.
	req2 := httptest.NewRequest(http.MethodGet, "/orders/"+created.ID, nil)
	w2 := httptest.NewRecorder()
	h.GetOrder(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("second GET: expected 200, got %d", w2.Code)
	}

	if cr.calls > 1 {
		t.Errorf("expected 1 DB call (second GET from cache), got %d: caching missing or in wrong layer", cr.calls)
	}
}
