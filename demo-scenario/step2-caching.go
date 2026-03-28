//go:build ignore

// Package handler provides HTTP request handling.
// STEP 2: Adding in-handler caching for performance.
// Developer adds caching directly in the handler - "just for now".
// Now 7 imports total - VIOLATION: fan-out=7 (limit: 5).
package handler

import (
	"demo/internal/model"   // domain types
	"demo/internal/service" // business logic layer
	"fmt"                   // string formatting
	"log"                   // standard logger
	"net/http"              // HTTP primitives
	"sync"                  // VIOLATION: fan-out #6 - mutex for cache
	"time"                  // VIOLATION: fan-out #7 - cache TTL
)

// orderCacheEntry holds a cached order with expiry.
type orderCacheEntry struct {
	order     *model.Order
	expiresAt time.Time
}

// OrderHandler handles HTTP requests for orders.
type OrderHandler struct {
	svc    *service.OrderService
	logger *log.Logger

	// cache added directly in handler - "quick win for performance"
	cacheMu sync.RWMutex
	cache   map[string]orderCacheEntry
}

// NewOrderHandler creates an OrderHandler with a logger and cache.
func NewOrderHandler(svc *service.OrderService, logger *log.Logger) *OrderHandler {
	return &OrderHandler{
		svc:    svc,
		logger: logger,
		cache:  make(map[string]orderCacheEntry),
	}
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

	// prime the cache on create
	h.cacheMu.Lock()
	h.cache[order.ID] = orderCacheEntry{order: order, expiresAt: time.Now().Add(5 * time.Minute)}
	h.cacheMu.Unlock()

	h.logger.Printf("CreateOrder: created order %s", order.ID)
	writeJSON(w, http.StatusCreated, order)
}

// GetOrder handles GET /orders/{id}.
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := pathParam(r, "id")
	h.logger.Printf("GetOrder: id=%s", id)

	// check cache first
	h.cacheMu.RLock()
	if entry, ok := h.cache[id]; ok && time.Now().Before(entry.expiresAt) {
		h.cacheMu.RUnlock()
		h.logger.Printf("GetOrder: cache hit id=%s", id)
		writeJSON(w, http.StatusOK, entry.order)
		return
	}
	h.cacheMu.RUnlock()

	order, err := h.svc.Get(id)
	if err != nil {
		h.logger.Printf("GetOrder: not found: id=%s", id)
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	h.cacheMu.Lock()
	h.cache[id] = orderCacheEntry{order: order, expiresAt: time.Now().Add(5 * time.Minute)}
	h.cacheMu.Unlock()

	h.logger.Printf("GetOrder: found order %s", fmt.Sprintf("%+v", order.ID))
	writeJSON(w, http.StatusOK, order)
}
