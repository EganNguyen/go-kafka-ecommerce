package grpc

import (
	"context"

	"github.com/egannguyen/go-kafka-ecommerce/currency-service/internal/delivery/grpc/pb"
	"github.com/egannguyen/go-kafka-ecommerce/currency-service/internal/domain"
)

type Server struct {
	pb.UnimplementedCurrencyServiceServer
	useCase domain.CurrencyService
}

func NewServer(useCase domain.CurrencyService) *Server {
	return &Server{useCase: useCase}
}

func (s *Server) GetSupportedCurrencies(ctx context.Context, req *pb.Empty) (*pb.GetSupportedCurrenciesResponse, error) {
	codes, err := s.useCase.GetSupportedCurrencies(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.GetSupportedCurrenciesResponse{CurrencyCodes: codes}, nil
}

func (s *Server) Convert(ctx context.Context, req *pb.CurrencyConversionRequest) (*pb.Money, error) {
	from := domain.Money{
		CurrencyCode: req.From.CurrencyCode,
		Units:        req.From.Units,
		Nanos:        req.From.Nanos,
	}

	converted, err := s.useCase.Convert(ctx, from, req.ToCode)
	if err != nil {
		return nil, err
	}

	return &pb.Money{
		CurrencyCode: converted.CurrencyCode,
		Units:        converted.Units,
		Nanos:        converted.Nanos,
	}, nil
}
