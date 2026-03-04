package domain

import "context"

type CurrencyService interface {
	Convert(ctx context.Context, from Money, toCode string) (Money, error)
}
