//go:build ignore

// Package handler provides HTTP request handling.
// STEP 4: Fixed version - facade extracted.
// OrderCache and OrderMetrics moved to their own types.
// Handler back to 4 imports - within fan-out budget (limit: 5).
package handler

import (
	"demo/internal/model"   // domain types
	"demo/internal/service" // business logic layer
	"log"                   // standard logger
	"net/http"              // HTTP primitives
)

// OrderCache is a narrow interface - handler does not know the implementation.
type OrderCache interface {
	Get(id string) (*model.Order, bool)
	Set(order *model.Order)
}

// OrderMetrics is a narrow interface for recording handler events.
type OrderMetrics interface {
	RecordCreate(err error)
	RecordGet(cacheHit bool)
}

// OrderHandler handles HTTP requests for orders.
// 4 imports, 3 dependencies - within all budgets.
type OrderHandler struct {
	svc     *service.OrderService
	logger  *log.Logger
	cache   OrderCache   // injected facade - no sync/time imports needed here
	metrics OrderMetrics // injected facade - no expvar imports needed here
}

// NewOrderHandler creates an OrderHandler.
func NewOrderHandler(
	svc *service.OrderService,
	logger *log.Logger,
	cache OrderCache,
	metrics OrderMetrics,
) *OrderHandler {
	return &OrderHandler{svc: svc, logger: logger, cache: cache, metrics: metrics}
}

// CreateOrder handles POST /orders.
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	h.logger.Println("CreateOrder called")

	var req model.CreateOrderRequest
	if err := decodeJSON(r, &req); err != nil {
		h.metrics.RecordCreate(err)
		h.logger.Printf("CreateOrder: decode error: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	order, err := h.svc.Create(req.UserID, req.Items)
	if err != nil {
		h.metrics.RecordCreate(err)
		h.logger.Printf("CreateOrder: service error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.metrics.RecordCreate(nil)
	h.cache.Set(order)
	h.logger.Printf("CreateOrder: created order %s", order.ID)
	writeJSON(w, http.StatusCreated, order)
}

// GetOrder handles GET /orders/{id}.
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := pathParam(r, "id")
	h.logger.Printf("GetOrder: id=%s", id)

	if order, ok := h.cache.Get(id); ok {
		h.metrics.RecordGet(true)
		h.logger.Printf("GetOrder: cache hit id=%s", id)
		writeJSON(w, http.StatusOK, order)
		return
	}

	order, err := h.svc.Get(id)
	if err != nil {
		h.logger.Printf("GetOrder: not found: id=%s", id)
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	h.metrics.RecordGet(false)
	h.cache.Set(order)
	h.logger.Printf("GetOrder: found order %s", order.ID)
	writeJSON(w, http.StatusOK, order)
}
