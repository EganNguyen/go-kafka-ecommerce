package domain

import "context"

type PaymentService interface {
	Charge(ctx context.Context, amount Money, card CreditCardInfo) (string, error)
}

type CreditCardInfo struct {
	Number          string
	CVV             int32
	ExpirationMonth int32
	ExpirationYear  int32
}
