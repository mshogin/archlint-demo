//go:build ignore

// Package handler provides HTTP request handling.
// STEP 1: Adding structured logging.
// Developer adds logging for observability - makes sense.
// Now 5 imports total - still within budget (fan-out limit: 5).
package handler

import (
	"demo/internal/model"   // domain types
	"demo/internal/service" // business logic layer
	"fmt"                   // string formatting for log messages
	"log"                   // standard logger
	"net/http"              // HTTP primitives
)

// OrderHandler handles HTTP requests for orders.
type OrderHandler struct {
	svc    *service.OrderService
	logger *log.Logger
}

// NewOrderHandler creates an OrderHandler with a logger.
func NewOrderHandler(svc *service.OrderService, logger *log.Logger) *OrderHandler {
	return &OrderHandler{svc: svc, logger: logger}
}

// CreateOrder handles POST /orders.
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	h.logger.Println("CreateOrder called")

	var req model.CreateOrderRequest
	if err := decodeJSON(r, &req); err != nil {
		h.logger.Printf("CreateOrder: decode error: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	order, err := h.svc.Create(req.UserID, req.Items)
	if err != nil {
		h.logger.Printf("CreateOrder: service error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Printf("CreateOrder: created order %s", order.ID)
	writeJSON(w, http.StatusCreated, order)
}

// GetOrder handles GET /orders/{id}.
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := pathParam(r, "id")
	h.logger.Printf("GetOrder: id=%s", id)

	order, err := h.svc.Get(id)
	if err != nil {
		h.logger.Printf("GetOrder: not found: id=%s", id)
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	h.logger.Printf("GetOrder: found order %s", fmt.Sprintf("%+v", order.ID))
	writeJSON(w, http.StatusOK, order)
}
