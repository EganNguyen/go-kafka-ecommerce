package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/egannguyen/go-kafka-ecommerce/cart-service/internal/usecase"
)

// Handler handles HTTP requests for the cart service.
type Handler struct {
	cartUseCase usecase.CartUseCase
}

func NewHandler(cartUseCase usecase.CartUseCase) *Handler {
	return &Handler{
		cartUseCase: cartUseCase,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/cart/{id}", h.handleGetCart)
	mux.HandleFunc("POST /api/cart/{id}/items", h.handleAddItemToCart)
}

func (h *Handler) handleGetCart(w http.ResponseWriter, r *http.Request) {
	cartID := r.PathValue("id")
	if cartID == "" {
		http.Error(w, "missing cart id", http.StatusBadRequest)
		return
	}

	cart, err := h.cartUseCase.GetCart(r.Context(), cartID)
	if err != nil {
		slog.Error("Failed to get cart", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cart.Items)
}

type AddCartItemRequest struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

func (h *Handler) handleAddItemToCart(w http.ResponseWriter, r *http.Request) {
	cartID := r.PathValue("id")
	if cartID == "" {
		http.Error(w, "missing cart id", http.StatusBadRequest)
		return
	}

	var req AddCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.cartUseCase.AddItemToCart(r.Context(), cartID, req.ProductID, req.Quantity, req.Price); err != nil {
		slog.Error("Failed to add item to cart", "err", err)
		http.Error(w, "failed to add item to cart", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// EnableCORS is a middleware to allow the React frontend to connect.
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
