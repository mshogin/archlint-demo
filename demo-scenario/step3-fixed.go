//go:build ignore

// Package handler provides HTTP request handling.
// STEP 3: Fixed - cache logic moved into service layer.
//
// The fix: OrderService.GetWithCache and OrderService.CreateAndCache
// encapsulate the cache interaction. Handler only talks to service.
// Repo stays invisible to handler.
//
// archlint: OK  handler/order.go  0 violations
// Handler -> Service -> Repo.  Correct direction restored.
package handler

import (
	"demo/internal/model"   // domain types
	"demo/internal/service" // business logic layer - single allowed downstream dependency
	"net/http"              // HTTP primitives
)

// OrderHandler handles HTTP requests for orders.
// Back to a single downstream dependency: service.OrderService.
// No repo imports. No direct cache access. Clean layer boundary.
type OrderHandler struct {
	svc *service.OrderService
}

// NewOrderHandler creates an OrderHandler.
func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// GetOrder handles GET /orders/{id}.
// Cache check now lives in service.GetWithCache - handler does not care how.
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := pathParam(r, "id")

	// Service encapsulates: cache lookup -> DB fallback -> cache populate.
	// Handler knows nothing about repos or caches.
	order, err := h.svc.GetWithCache(id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, order)
}

// CreateOrder handles POST /orders.
// Side effects (cache population, user counters) are service responsibilities.
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req model.CreateOrderRequest
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Service handles: validation -> persist -> cache populate -> counters.
	// One call. Handler stays thin.
	order, err := h.svc.Create(req.UserID, req.Items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, order)
}
