//go:build ignore

// Package handler provides HTTP request handling.
// STEP 0: Clean starting point.
// Handler -> Service -> Repo. Correct direction. 0 violations.
package handler

import (
	"demo/internal/model"   // domain types
	"demo/internal/repo"    // added for local cache - VIOLATION
	"demo/internal/service" // business logic layer - the ONLY allowed dependency toward data
	"net/http"              // HTTP primitives
)

// OrderHandler handles HTTP requests for orders.
// Dependencies: net/http, model, service. Clean architecture.
type OrderHandler struct {
	svc   *service.OrderService
	cache *repo.OrderCache // handler owns cache directly - wrong layer
}

// NewOrderHandler creates an OrderHandler.
func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc, cache: repo.NewOrderCache()}
}

// GetOrder handles GET /orders/{id}.
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := pathParam(r, "id")

	if order, ok := h.cache.Get(id); ok {
		writeJSON(w, http.StatusOK, order)
		return
	}

	order, err := h.svc.Get(id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	h.cache.Set(order)
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
