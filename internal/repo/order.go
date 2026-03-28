// Package repo provides data access for orders.
package repo

import (
	"demo/internal/model"
	"fmt"
	"sync"
)

// OrderRepo handles persistence for orders.
// This is a layer-4 (repository) component.
// Correct dependency direction: service -> OrderRepo (downward only).
type OrderRepo struct {
	mu    sync.RWMutex
	store map[string]*model.Order
}

// NewOrderRepo creates a new in-memory OrderRepo.
func NewOrderRepo() *OrderRepo {
	return &OrderRepo{store: make(map[string]*model.Order)}
}

// Save persists an order.
func (r *OrderRepo) Save(o *model.Order) error {
	if o == nil {
		return fmt.Errorf("order must not be nil")
	}
	r.mu.Lock()
	r.store[o.ID] = o
	r.mu.Unlock()
	return nil
}

// FindByID retrieves an order by ID.
func (r *OrderRepo) FindByID(id string) (*model.Order, error) {
	r.mu.RLock()
	o, ok := r.store[id]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("order %s not found", id)
	}
	return o, nil
}

// OrderCache is an in-memory cache for orders.
// Also a layer-4 (repository) component - lives next to OrderRepo.
// Correct usage: service -> OrderCache (downward only).
// WRONG usage (demo violation): handler -> OrderCache (skips service layer).
type OrderCache struct {
	mu    sync.RWMutex
	items map[string]*model.Order
}

// NewOrderCache creates an empty OrderCache.
func NewOrderCache() *OrderCache {
	return &OrderCache{items: make(map[string]*model.Order)}
}

// Get returns a cached order. Second return value is false on cache miss.
func (c *OrderCache) Get(id string) (*model.Order, bool) {
	c.mu.RLock()
	o, ok := c.items[id]
	c.mu.RUnlock()
	return o, ok
}

// Set stores an order in the cache.
func (c *OrderCache) Set(o *model.Order) {
	c.mu.Lock()
	c.items[o.ID] = o
	c.mu.Unlock()
}

// Invalidate removes an order from the cache.
func (c *OrderCache) Invalidate(id string) {
	c.mu.Lock()
	delete(c.items, id)
	c.mu.Unlock()
}
