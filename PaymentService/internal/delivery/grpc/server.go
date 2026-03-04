package grpc

import (
	"context"

	"github.com/egannguyen/go-kafka-ecommerce/payment-service/internal/delivery/grpc/pb"
	"github.com/egannguyen/go-kafka-ecommerce/payment-service/internal/domain"
)

type Server struct {
	pb.UnimplementedPaymentServiceServer
	useCase domain.PaymentService
}

func NewServer(useCase domain.PaymentService) *Server {
	return &Server{useCase: useCase}
}

func (s *Server) Charge(ctx context.Context, req *pb.ChargeRequest) (*pb.ChargeResponse, error) {
	card := domain.CreditCardInfo{
		Number:          req.CreditCard.Number,
		CVV:             req.CreditCard.Cvv,
		ExpirationMonth: req.CreditCard.ExpirationMonth,
		ExpirationYear:  req.CreditCard.ExpirationYear,
	}

	amount := domain.Money{
		CurrencyCode: req.Amount.CurrencyCode,
		Units:        req.Amount.Units,
		Nanos:        req.Amount.Nanos,
	}

	txID, err := s.useCase.Charge(ctx, amount, card)
	if err != nil {
		return nil, err
	}

	return &pb.ChargeResponse{TransactionId: txID}, nil
}
