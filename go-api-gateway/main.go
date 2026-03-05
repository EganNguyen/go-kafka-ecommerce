package main

import (
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func main() {
	slog.Info("Starting API Gateway...")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// Configure routes
	setupProxy(mux, "/api/products", "http://productcatalog-service:8080")
	setupProxy(mux, "/api/products/", "http://productcatalog-service:8080")
	setupProxy(mux, "/api/orders", "http://checkout-service:8080")
	setupProxy(mux, "/api/orders/", "http://checkout-service:8080")
	setupProxy(mux, "/api/cart", "http://cart-service:8080")
	setupProxy(mux, "/api/cart/", "http://cart-service:8080")

	// Apply CORS middleware
	handler := enableCORS(mux)

	slog.Info("API Gateway listening on port " + port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("API Gateway failed: %v", err)
	}
}

func setupProxy(mux *http.ServeMux, path, target string) {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Invalid target URL for path %s: %v", path, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Update the request headers
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	}

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Proxying request", "method", r.Method, "path", r.URL.Path, "target", target)
		proxy.ServeHTTP(w, r)
	})
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
