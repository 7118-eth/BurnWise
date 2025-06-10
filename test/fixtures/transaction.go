package fixtures

import (
	"time"

	"budget-tracker/internal/models"
)

type TransactionBuilder struct {
	tx *models.Transaction
}

func NewTransaction() *TransactionBuilder {
	return &TransactionBuilder{
		tx: &models.Transaction{
			Type:        models.TransactionTypeExpense,
			Amount:      100.00,
			Currency:    "USD",
			AmountUSD:   100.00,
			CategoryID:  1,
			Description: "Test transaction",
			Date:        time.Now(),
		},
	}
}

func (b *TransactionBuilder) WithType(txType models.TransactionType) *TransactionBuilder {
	b.tx.Type = txType
	return b
}

func (b *TransactionBuilder) WithAmount(amount float64) *TransactionBuilder {
	b.tx.Amount = amount
	b.tx.AmountUSD = amount
	return b
}

func (b *TransactionBuilder) WithCurrency(currency string) *TransactionBuilder {
	b.tx.Currency = currency
	return b
}

func (b *TransactionBuilder) WithAmountUSD(amountUSD float64) *TransactionBuilder {
	b.tx.AmountUSD = amountUSD
	return b
}

func (b *TransactionBuilder) WithCategory(categoryID uint) *TransactionBuilder {
	b.tx.CategoryID = categoryID
	return b
}

func (b *TransactionBuilder) WithDescription(description string) *TransactionBuilder {
	b.tx.Description = description
	return b
}

func (b *TransactionBuilder) WithDate(date time.Time) *TransactionBuilder {
	b.tx.Date = date
	return b
}

func (b *TransactionBuilder) Build() *models.Transaction {
	return b.tx
}

func CreateSampleTransactions() []*models.Transaction {
	now := time.Now()
	
	return []*models.Transaction{
		NewTransaction().
			WithType(models.TransactionTypeIncome).
			WithAmount(5000).
			WithCategory(1).
			WithDescription("Monthly salary").
			WithDate(now.AddDate(0, 0, -1)).
			Build(),
		NewTransaction().
			WithType(models.TransactionTypeExpense).
			WithAmount(1500).
			WithCategory(5).
			WithDescription("October rent").
			WithDate(now.AddDate(0, 0, -2)).
			Build(),
		NewTransaction().
			WithType(models.TransactionTypeExpense).
			WithAmount(250).
			WithCategory(6).
			WithDescription("Weekly groceries").
			WithDate(now).
			Build(),
		NewTransaction().
			WithType(models.TransactionTypeExpense).
			WithAmount(50).
			WithCategory(7).
			WithDescription("Gas").
			WithDate(now).
			Build(),
	}
}