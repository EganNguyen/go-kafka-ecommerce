package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/egannguyen/go-kafka-ecommerce/payment-service/internal/domain"
)

type transactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) domain.TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Save(ctx context.Context, tx *domain.Transaction) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO transactions (id, amount_units, amount_nanos, currency_code, card_number_last4) VALUES ($1, $2, $3, $4, $5)",
		tx.ID, tx.AmountUnits, tx.AmountNanos, tx.CurrencyCode, tx.CardNumberLast4,
	)
	if err != nil {
		return fmt.Errorf("failed to save transaction: %w", err)
	}
	return nil
}
