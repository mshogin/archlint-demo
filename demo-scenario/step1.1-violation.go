//go:build ignore

package handler

import (
	"demo/internal/model"
	"demo/internal/repo"
	"demo/internal/service"
	"net/http"
)

type OrderHandler struct {
	svc   *service.OrderService
	cache *repo.OrderCache
}

func NewOrderHandler(svc *service.OrderService, cache *repo.OrderCache) *OrderHandler {
	return &OrderHandler{svc: svc, cache: cache}
}

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

	writeJSON(w, http.StatusOK, order)
}

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
