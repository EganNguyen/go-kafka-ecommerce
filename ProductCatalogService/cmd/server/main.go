package main

import (
	"context"
	"log"
	"log/slog"
	stdhttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	deliveryGrpc "github.com/egannguyen/go-kafka-ecommerce/product-catalog-service/internal/delivery/grpc"
	"github.com/egannguyen/go-kafka-ecommerce/product-catalog-service/internal/delivery/grpc/pb"
	deliveryHttp "github.com/egannguyen/go-kafka-ecommerce/product-catalog-service/internal/delivery/http"
	"github.com/egannguyen/go-kafka-ecommerce/product-catalog-service/internal/infrastructure/persistence/mongodb"
	"github.com/egannguyen/go-kafka-ecommerce/product-catalog-service/internal/usecase"
	stdgrpc "google.golang.org/grpc"
	"net"
)

func main() {
	mongoURL := os.Getenv("MONGODB_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://mongodb:27017"
	}

	db, err := mongodb.InitDB(mongoURL)
	if err != nil {
		log.Fatal("Failed to connect to mongodb:", err)
	}

	repo := mongodb.NewProductRepository(db)
	catalogUseCase := usecase.NewCatalogUseCase(repo)
	handler := deliveryHttp.NewHandler(catalogUseCase)

	// HTTP Handler
	mux := stdhttp.NewServeMux()
	handler.RegisterRoutes(mux)

	srv := &stdhttp.Server{
		Addr:    ":8080",
		Handler: deliveryHttp.EnableCORS(mux),
	}

	// gRPC Server
	grpcSrv := stdgrpc.NewServer()
	pb.RegisterProductCatalogServiceServer(grpcSrv, deliveryGrpc.NewServer(catalogUseCase))

	go func() {
		slog.Info("ProductCatalogService HTTP starting on :8080")
		if err := srv.ListenAndServe(); err != nil && err != stdhttp.ErrServerClosed {
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
