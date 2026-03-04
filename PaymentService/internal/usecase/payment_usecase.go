package usecase

import (
	"context"
	"fmt"

	"github.com/egannguyen/go-kafka-ecommerce/payment-service/internal/domain"
	"github.com/google/uuid"
)

type paymentUseCase struct {
	repo domain.TransactionRepository
}

func NewPaymentUseCase(repo domain.TransactionRepository) domain.PaymentService {
	return &paymentUseCase{repo: repo}
}

func (u *paymentUseCase) Charge(ctx context.Context, amount domain.Money, card domain.CreditCardInfo) (string, error) {
	if len(card.Number) < 13 {
		return "", fmt.Errorf("invalid credit card number length")
	}

	transactionID := uuid.New().String()

	// Persist transaction
	last4 := ""
	if len(card.Number) >= 4 {
		last4 = card.Number[len(card.Number)-4:]
	}

	tx := &domain.Transaction{
		ID:              transactionID,
		AmountUnits:     amount.Units,
		AmountNanos:     amount.Nanos,
		CurrencyCode:    amount.CurrencyCode,
		CardNumberLast4: last4,
	}

	if err := u.repo.Save(ctx, tx); err != nil {
		return "", fmt.Errorf("failed to persist transaction: %w", err)
	}

	return transactionID, nil
}
