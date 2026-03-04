package main

import (
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/egannguyen/go-kafka-ecommerce/currency-service/internal/delivery/grpc"
	"github.com/egannguyen/go-kafka-ecommerce/currency-service/internal/delivery/grpc/pb"
	"github.com/egannguyen/go-kafka-ecommerce/currency-service/internal/usecase"
	g "google.golang.org/grpc"
)

func main() {
	useCase := usecase.NewCurrencyUseCase()
	grpcSrv := g.NewServer()
	pb.RegisterCurrencyServiceServer(grpcSrv, grpc.NewServer(useCase))

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		slog.Info("CurrencyService gRPC starting on :50051")
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down CurrencyService...")
	grpcSrv.GracefulStop()
	slog.Info("CurrencyService exited")
}
