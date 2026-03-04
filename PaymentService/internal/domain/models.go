package domain

import "context"

type CreditCardInfo struct {
	Number          string
	CVV             int32
	ExpirationMonth int32
	ExpirationYear  int32
}

type Money struct {
	CurrencyCode string
	Units        int64
	Nanos        int32
}

type PaymentService interface {
	Charge(ctx context.Context, amount Money, card CreditCardInfo) (string, error)
}
