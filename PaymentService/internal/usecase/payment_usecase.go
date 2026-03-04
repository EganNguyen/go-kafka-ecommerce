package usecase

import (
	"context"
	"fmt"

	"github.com/egannguyen/go-kafka-ecommerce/payment-service/internal/domain"
	"github.com/google/uuid"
)

type paymentUseCase struct{}

func NewPaymentUseCase() domain.PaymentService {
	return &paymentUseCase{}
}

func (u *paymentUseCase) Charge(ctx context.Context, amount domain.Money, card domain.CreditCardInfo) (string, error) {
	if len(card.Number) < 13 {
		return "", fmt.Errorf("invalid credit card number length")
	}

	// Mocking successful payment
	transactionID := uuid.New().String()
	return transactionID, nil
}
