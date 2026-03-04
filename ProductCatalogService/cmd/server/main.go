package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/egannguyen/go-kafka-ecommerce/product-catalog-service/internal/delivery/grpc"
	"github.com/egannguyen/go-kafka-ecommerce/product-catalog-service/internal/delivery/grpc/pb"
	"github.com/egannguyen/go-kafka-ecommerce/product-catalog-service/internal/delivery/http"
	"github.com/egannguyen/go-kafka-ecommerce/product-catalog-service/internal/infrastructure/persistence/postgres"
	"github.com/egannguyen/go-kafka-ecommerce/product-catalog-service/internal/usecase"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"net"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://ecommerce:ecommerce@postgres:5432/ecommerce?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Wait for DB to be ready
	for i := 0; i < 10; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}

	repo := postgres.NewProductRepository(db)
	catalogUseCase := usecase.NewCatalogUseCase(repo)
	handler := http.NewHandler(catalogUseCase)

	// HTTP Handler
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: http.EnableCORS(mux),
	}

	// gRPC Server
	grpcSrv := grpc.NewServer()
	pb.RegisterProductCatalogServiceServer(grpcSrv, grpc.NewServer(catalogUseCase))

	go func() {
		slog.Info("ProductCatalogService HTTP starting on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http listen: %s\n", err)
		}
	}()

	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("grpc failed to listen: %v", err)
		}
		slog.Info("ProductCatalogService gRPC starting on :50051")
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatalf("grpc failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	slog.Info("Server exiting")
}
