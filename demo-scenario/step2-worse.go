//go:build ignore

// Package handler provides HTTP request handling.
// STEP 2: Making it worse - second direct repo import added.
//
// The story: "While I'm here, I'll also invalidate the user's order list cache
// on CreateOrder. The user cache is in repo too. One more import - no big deal."
//
// Now two forbidden dependencies from handler layer:
//   handler -> repo.OrderCache   (layer skip #1)
//   handler -> repo.UserRepo     (layer skip #2)
//
// archlint: VIOLATION  handler/order.go  layer=handler -> repo  (2 forbidden deps)
// This is how "just this once" becomes a pattern.
package handler

import (
	"demo/internal/model"   // domain types
	"demo/internal/repo"    // VIOLATION: handler -> repo (2 forbidden dependencies now)
	"demo/internal/service" // business logic layer
	"net/http"              // HTTP primitives
)

// OrderHandler handles HTTP requests for orders.
// Two direct repo dependencies - the violation is spreading.
type OrderHandler struct {
	svc      *service.OrderService
	cache    *repo.OrderCache // VIOLATION: layer skip #1
	userRepo *repo.UserRepo   // VIOLATION: layer skip #2 - "just needed user lookup"
}

// NewOrderHandler creates an OrderHandler.
// VIOLATION: two repo types injected - handler layer is leaking into data layer.
func NewOrderHandler(
	svc *service.OrderService,
	cache *repo.OrderCache,
	userRepo *repo.UserRepo,
) *OrderHandler {
	return &OrderHandler{svc: svc, cache: cache, userRepo: userRepo}
}

// GetOrder handles GET /orders/{id}.
// VIOLATION: direct cache lookup in handler.
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := pathParam(r, "id")

	// VIOLATION: cache check belongs in service.GetWithCache, not here.
	if order, ok := h.cache.Get(id); ok {
		writeJSON(w, http.StatusOK, order)
		return
	}

	order, err := h.svc.Get(id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// VIOLATION: handler populates cache directly after service call.
	h.cache.Set(order)
	writeJSON(w, http.StatusOK, order)
}

// CreateOrder handles POST /orders.
// VIOLATION: handler reaches into repo to invalidate user cache after order creation.
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req model.CreateOrderRequest
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	order, err := h.svc.Create(req.UserID, req.Items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// VIOLATION: "quick" post-create side effects directly in handler.
	// userRepo and cache belong in the service layer.
	h.cache.Set(order)
	_ = h.userRepo // touch - user order count cache invalidation "coming soon"

	writeJSON(w, http.StatusCreated, order)
}
