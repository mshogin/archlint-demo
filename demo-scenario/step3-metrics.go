//go:build ignore

// Package handler provides HTTP request handling.
// STEP 3: Adding request metrics (latency tracking).
// Developer adds metrics inline - one more "quick addition".
// Now 8 imports total - VIOLATION: fan-out=8 (limit: 5), getting worse.
package handler

import (
	"demo/internal/model"   // domain types
	"demo/internal/service" // business logic layer
	"fmt"                   // string formatting
	"log"                   // standard logger
	"net/http"              // HTTP primitives
	"sync"                  // VIOLATION: fan-out #6 - mutex for cache
	"time"                  // VIOLATION: fan-out #7 - cache TTL and latency
	"expvar"                // VIOLATION: fan-out #8 - metrics counters
)

// orderCacheEntry holds a cached order with expiry.
type orderCacheEntry struct {
	order     *model.Order
	expiresAt time.Time
}

var (
	createOrderCount = expvar.NewInt("order.create.count")
	createOrderErrors = expvar.NewInt("order.create.errors")
	getOrderCount    = expvar.NewInt("order.get.count")
	getOrderCacheHit = expvar.NewInt("order.get.cache_hit")
)

// OrderHandler handles HTTP requests for orders.
type OrderHandler struct {
	svc    *service.OrderService
	logger *log.Logger

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
	start := time.Now()
	createOrderCount.Add(1)
	h.logger.Println("CreateOrder called")

	var req model.CreateOrderRequest
	if err := decodeJSON(r, &req); err != nil {
		createOrderErrors.Add(1)
		h.logger.Printf("CreateOrder: decode error: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	order, err := h.svc.Create(req.UserID, req.Items)
	if err != nil {
		createOrderErrors.Add(1)
		h.logger.Printf("CreateOrder: service error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.cacheMu.Lock()
	h.cache[order.ID] = orderCacheEntry{order: order, expiresAt: time.Now().Add(5 * time.Minute)}
	h.cacheMu.Unlock()

	h.logger.Printf("CreateOrder: created order %s in %v", order.ID, time.Since(start))
	writeJSON(w, http.StatusCreated, order)
}

// GetOrder handles GET /orders/{id}.
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	getOrderCount.Add(1)
	id := pathParam(r, "id")
	h.logger.Printf("GetOrder: id=%s", id)

	h.cacheMu.RLock()
	if entry, ok := h.cache[id]; ok && time.Now().Before(entry.expiresAt) {
		h.cacheMu.RUnlock()
		getOrderCacheHit.Add(1)
		h.logger.Printf("GetOrder: cache hit id=%s in %v", id, time.Since(start))
		writeJSON(w, http.StatusOK, entry.order)
		return
	}
	h.cacheMu.RUnlock()

	order, err := h.svc.Get(id)
	if err != nil {
		h.logger.Printf("GetOrder: not found: id=%s in %v", id, time.Since(start))
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	h.cacheMu.Lock()
	h.cache[id] = orderCacheEntry{order: order, expiresAt: time.Now().Add(5 * time.Minute)}
	h.cacheMu.Unlock()

	h.logger.Printf("GetOrder: found order %s in %v", fmt.Sprintf("%+v", order.ID), time.Since(start))
	writeJSON(w, http.StatusOK, order)
}
