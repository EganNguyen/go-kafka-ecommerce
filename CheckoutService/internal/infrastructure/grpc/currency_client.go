package grpc

import (
	"context"
	"fmt"

	"github.com/egannguyen/go-kafka-ecommerce/checkout-service/internal/delivery/grpc/pb"
	"github.com/egannguyen/go-kafka-ecommerce/checkout-service/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type currencyServiceClient struct {
	client pb.CurrencyServiceClient
}

func NewCurrencyServiceClient(addr string) (domain.CurrencyService, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to currency service: %w", err)
	}

	client := pb.NewCurrencyServiceClient(conn)
	return &currencyServiceClient{client: client}, nil
}

func (s *currencyServiceClient) Convert(ctx context.Context, from domain.Money, toCode string) (domain.Money, error) {
	resp, err := s.client.Convert(ctx, &pb.CurrencyConversionRequest{
		From: &pb.Money{
			CurrencyCode: from.CurrencyCode,
			Units:        from.Units,
			Nanos:        from.Nanos,
		},
		ToCode: toCode,
	})
	if err != nil {
		return domain.Money{}, err
	}

	return domain.Money{
		CurrencyCode: resp.CurrencyCode,
		Units:        resp.Units,
		Nanos:        resp.Nanos,
	}, nil
}
