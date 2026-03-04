package grpc

import (
	"context"
	"fmt"

	"github.com/egannguyen/go-kafka-ecommerce/checkout-service/internal/delivery/grpc/pb"
	"github.com/egannguyen/go-kafka-ecommerce/checkout-service/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type paymentServiceClient struct {
	client pb.PaymentServiceClient
}

func NewPaymentServiceClient(addr string) (domain.PaymentService, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to payment service: %w", err)
	}

	client := pb.NewPaymentServiceClient(conn)
	return &paymentServiceClient{client: client}, nil
}

func (s *paymentServiceClient) Charge(ctx context.Context, amount domain.Money, card domain.CreditCardInfo) (string, error) {
	resp, err := s.client.Charge(ctx, &pb.ChargeRequest{
		Amount: &pb.Money{
			CurrencyCode: amount.CurrencyCode,
			Units:        amount.Units,
			Nanos:        amount.Nanos,
		},
		CreditCard: &pb.CreditCardInfo{
			Number:          card.Number,
			Cvv:             card.CVV,
			ExpirationMonth: card.ExpirationMonth,
			ExpirationYear:  card.ExpirationYear,
		},
	})
	if err != nil {
		return "", err
	}

	return resp.TransactionId, nil
}
