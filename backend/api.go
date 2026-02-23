package main

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

// API holds the HTTP handlers and dependencies.
type API struct {
	db         *sql.DB
	placeOrder func(cmd *PlaceOrder) error
}

// enableCORS is middleware to allow the React frontend to connect.
func enableCORS(next http.Handler) http.Handler {
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

func (a *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/products", a.handleGetProducts)
	mux.HandleFunc("POST /api/orders", a.handleCreateOrder)
	mux.HandleFunc("GET /api/orders", a.handleGetOrders)
}

func (a *API) handleGetProducts(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.QueryContext(r.Context(), "SELECT id, name, description, price, image_url, category, stock FROM products ORDER BY name")
	if err != nil {
		http.Error(w, "failed to query products", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL, &p.Category, &p.Stock); err != nil {
			http.Error(w, "failed to scan product", http.StatusInternalServerError)
			return
		}
		products = append(products, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

type CreateOrderRequest struct {
	Items []OrderItem `json:"items"`
}

func (a *API) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Items) == 0 {
		http.Error(w, "order must have at least one item", http.StatusBadRequest)
		return
	}

	orderID := uuid.New().String()

	cmd := &PlaceOrder{
		OrderID: orderID,
		Items:   req.Items,
	}

	if err := a.placeOrder(cmd); err != nil {
		slog.Error("Failed to place order", "err", err)
		http.Error(w, "failed to place order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"order_id": orderID,
		"status":   "placed",
	})
}

func (a *API) handleGetOrders(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.QueryContext(r.Context(), "SELECT id, total_price, status, created_at FROM orders ORDER BY created_at DESC LIMIT 50")
	if err != nil {
		http.Error(w, "failed to query orders", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.TotalPrice, &o.Status, &o.CreatedAt); err != nil {
			http.Error(w, "failed to scan order", http.StatusInternalServerError)
			return
		}
		orders = append(orders, o)
	}

	// For each order, fetch its items.
	for i := range orders {
		itemRows, err := a.db.QueryContext(r.Context(),
			"SELECT product_id, name, price, quantity FROM order_items WHERE order_id = $1",
			orders[i].ID,
		)
		if err != nil {
			http.Error(w, "failed to query order items", http.StatusInternalServerError)
			return
		}
		defer itemRows.Close()

		for itemRows.Next() {
			var item OrderItem
			if err := itemRows.Scan(&item.ProductID, &item.Name, &item.Price, &item.Quantity); err != nil {
				http.Error(w, "failed to scan order item", http.StatusInternalServerError)
				return
			}
			orders[i].Items = append(orders[i].Items, item)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}
