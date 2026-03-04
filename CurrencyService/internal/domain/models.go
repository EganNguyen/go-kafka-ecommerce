package domain

import "context"

type Money struct {
	CurrencyCode string
	Units        int64
	Nanos        int32
}

type CurrencyService interface {
	Convert(ctx context.Context, from Money, toCode string) (Money, error)
	GetSupportedCurrencies(ctx context.Context) ([]string, error)
}
