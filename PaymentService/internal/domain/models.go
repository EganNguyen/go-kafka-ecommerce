package domain

import "context"

type CreditCardInfo struct {
	Number          string
	CVV             int32
	ExpirationMonth int32
	ExpirationYear  int32
}

type Transaction struct {
	ID            string
	AmountUnits   int64
	AmountNanos   int32
	CurrencyCode  string
	CardNumberLast4 string
	CreatedAt     string
}

type Money struct {
	CurrencyCode string
	Units        int64
	Nanos        int32
}

type PaymentService interface {
	Charge(ctx context.Context, amount Money, card CreditCardInfo) (string, error)
}

type TransactionRepository interface {
	Save(ctx context.Context, tx *Transaction) error
}
