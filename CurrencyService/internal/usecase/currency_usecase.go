package usecase

import (
	"context"
	"fmt"
	"math"

	"github.com/egannguyen/go-kafka-ecommerce/currency-service/internal/domain"
)

var exchangeRates = map[string]float64{
	"USD": 1.0,
	"EUR": 0.92,
	"GBP": 0.79,
	"JPY": 150.0,
}

type currencyUseCase struct{}

func NewCurrencyUseCase() domain.CurrencyService {
	return &currencyUseCase{}
}

func (u *currencyUseCase) GetSupportedCurrencies(ctx context.Context) ([]string, error) {
	currencies := make([]string, 0, len(exchangeRates))
	for code := range exchangeRates {
		currencies = append(currencies, code)
	}
	return currencies, nil
}

func (u *currencyUseCase) Convert(ctx context.Context, from domain.Money, toCode string) (domain.Money, error) {
	fromRate, ok := exchangeRates[from.CurrencyCode]
	if !ok {
		return domain.Money{}, fmt.Errorf("unsupported source currency: %s", from.CurrencyCode)
	}
	toRate, ok := exchangeRates[toCode]
	if !ok {
		return domain.Money{}, fmt.Errorf("unsupported target currency: %s", toCode)
	}

	// Convert to USD base
	totalUnits := float64(from.Units) + float64(from.Nanos)/1e9
	usdAmount := totalUnits / fromRate

	// Convert to Target currency
	targetAmount := usdAmount * toRate

	units, nanos := math.Modf(targetAmount)
	return domain.Money{
		CurrencyCode: toCode,
		Units:        int64(units),
		Nanos:        int32(math.Round(nanos * 1e9)),
	}, nil
}
