//go:build ignore

// Package handler provides HTTP request handling.
// STEP 1: "Quick fix" - developer imports repo directly from handler.
//
// The story: "GetOrder is slow because it always hits the DB.
// I just need to check the cache first. The cache lives in repo.
// I'll import it directly - it's faster than adding a method to the service."
//
// VIOLATION: handler -> repo (forbidden dependency, must go through service)
// archlint: VIOLATION  handler/order.go  layer=handler -> repo  (forbidden: skip service layer)
package handler

import (
	"demo/internal/model"   // domain types
	"demo/internal/repo"    // VIOLATION: handler -> repo (layer skip! must go through service)
	"demo/internal/service" // business logic layer
	"net/http"              // HTTP primitives
)

// OrderHandler handles HTTP requests for orders.
// VIOLATION: holds a direct reference to *repo.OrderCache
// instead of depending on service.OrderService for all data access.
type OrderHandler struct {
	svc   *service.OrderService
	cache *repo.OrderCache // VIOLATION: repo imported directly into handler
}

// NewOrderHandler creates an OrderHandler with a direct cache dependency.
// VIOLATION: *repo.OrderCache injected here - handler now bypasses service layer.
func NewOrderHandler(svc *service.OrderService, cache *repo.OrderCache) *OrderHandler {
	return &OrderHandler{svc: svc, cache: cache}
}

// GetOrder handles GET /orders/{id}.
// "Quick fix": check cache before calling service.
// Looks harmless. archlint fires immediately on save.
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := pathParam(r, "id")

	// VIOLATION: handler calls repo directly, bypassing service layer.
	// Cache lookup logic belongs in service, not here.
	if order, ok := h.cache.Get(id); ok {
		writeJSON(w, http.StatusOK, order)
		return
	}

	order, err := h.svc.Get(id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, order)
}

// CreateOrder handles POST /orders.
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

	writeJSON(w, http.StatusCreated, order)
}
