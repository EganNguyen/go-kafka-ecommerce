package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/egannguyen/go-kafka-ecommerce/checkout-service/internal/domain"
	"github.com/egannguyen/go-kafka-ecommerce/checkout-service/internal/usecase"
	"github.com/google/uuid"
)

type Handler struct {
	checkoutUseCase usecase.CheckoutUseCase
}

func NewHandler(checkoutUseCase usecase.CheckoutUseCase) *Handler {
	return &Handler{
		checkoutUseCase: checkoutUseCase,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/products", h.handleGetProducts)
	mux.HandleFunc("POST /api/orders", h.handleCreateOrder)
	mux.HandleFunc("GET /api/orders", h.handleGetOrders)
}

func (h *Handler) handleGetProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.checkoutUseCase.GetProducts(r.Context())
	if err != nil {
		slog.Error("Failed to get products", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

type CreateOrderRequest struct {
	Items []domain.OrderItem `json:"items"`
}

func (h *Handler) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	cmd := &domain.PlaceOrder{
		OrderID: uuid.New().String(),
		Items:   req.Items,
	}

	if err := h.checkoutUseCase.PlaceOrder(r.Context(), cmd); err != nil {
		slog.Error("Failed to place order", "err", err)
		http.Error(w, "failed to place order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"order_id": cmd.OrderID,
		"status":   "placed",
	})
}

func (h *Handler) handleGetOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.checkoutUseCase.GetRecentOrders(r.Context(), 50)
	if err != nil {
		slog.Error("Failed to get orders", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
