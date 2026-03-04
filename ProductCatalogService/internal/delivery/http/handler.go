package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/egannguyen/go-kafka-ecommerce/product-catalog-service/internal/usecase"
)

type Handler struct {
	useCase usecase.CatalogUseCase
}

func NewHandler(useCase usecase.CatalogUseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/products", h.handleListProducts)
	mux.HandleFunc("GET /api/products/{id}", h.handleGetProduct)
}

func (h *Handler) handleListProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.useCase.ListProducts(r.Context())
	if err != nil {
		slog.Error("Failed to list products", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func (h *Handler) handleGetProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing product id", http.StatusBadRequest)
		return
	}

	product, err := h.useCase.GetProduct(r.Context(), id)
	if err != nil {
		slog.Error("Failed to get product", "err", err, "id", id)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if product == nil {
		http.Error(w, "product not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
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
