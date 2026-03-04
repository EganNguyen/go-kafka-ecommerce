package main

import (
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/egannguyen/go-kafka-ecommerce/payment-service/internal/delivery/grpc"
	"github.com/egannguyen/go-kafka-ecommerce/payment-service/internal/delivery/grpc/pb"
	"github.com/egannguyen/go-kafka-ecommerce/payment-service/internal/infrastructure/persistence/postgres"
	"github.com/egannguyen/go-kafka-ecommerce/payment-service/internal/usecase"
	g "google.golang.org/grpc"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://ecommerce:ecommerce@postgres:5432/ecommerce?sslmode=disable"
	}

	db, err := postgres.InitDB(dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	repo := postgres.NewTransactionRepository(db)
	useCase := usecase.NewPaymentUseCase(repo)
	grpcSrv := g.NewServer()
	pb.RegisterPaymentServiceServer(grpcSrv, grpc.NewServer(useCase))

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		slog.Info("PaymentService gRPC starting on :50051")
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down PaymentService...")
	grpcSrv.GracefulStop()
	slog.Info("PaymentService exited")
}
