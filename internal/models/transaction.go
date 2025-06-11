package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type TransactionType string

const (
	TransactionTypeIncome   TransactionType = "income"
	TransactionTypeExpense  TransactionType = "expense"
	TransactionTypeTransfer TransactionType = "transfer"
)

type Transaction struct {
	ID                     uint            `gorm:"primaryKey" json:"id"`
	Type                   TransactionType `gorm:"type:varchar(20);not null" json:"type"`
	Amount                 float64         `gorm:"not null" json:"amount"`
	Currency               string          `gorm:"type:varchar(3);not null" json:"currency"`
	AmountUSD              float64         `gorm:"not null" json:"amount_usd"`
	CategoryID             uint            `gorm:"not null" json:"category_id"`
	Description            string          `gorm:"type:varchar(255)" json:"description"`
	Date                   time.Time       `gorm:"not null" json:"date"`
	RecurringTransactionID *uint           `json:"recurring_transaction_id,omitempty"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`
	DeletedAt              gorm.DeletedAt  `gorm:"index" json:"deleted_at,omitempty"`

	Category             Category              `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	RecurringTransaction *RecurringTransaction `gorm:"foreignKey:RecurringTransactionID" json:"recurring_transaction,omitempty"`
}

func (t *Transaction) Validate() error {
	if t.Type != TransactionTypeIncome && t.Type != TransactionTypeExpense && t.Type != TransactionTypeTransfer {
		return errors.New("invalid transaction type")
	}

	if t.Amount <= 0 {
		return errors.New("amount must be positive")
	}

	if len(t.Currency) != 3 {
		return errors.New("currency must be a 3-letter ISO code")
	}

	if t.CategoryID == 0 {
		return errors.New("category is required")
	}

	if t.Date.IsZero() {
		return errors.New("date is required")
	}

	return nil
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if err := t.Validate(); err != nil {
		return err
	}

	if t.Date.IsZero() {
		t.Date = time.Now()
	}

	return nil
}

type TransactionFilter struct {
	Type       TransactionType
	CategoryID uint
	StartDate  time.Time
	EndDate    time.Time
	MinAmount  float64
	MaxAmount  float64
	Currency   string
	Search     string
}

type TransactionSummary struct {
	TotalIncome   float64
	TotalExpenses float64
	Balance       float64
	Count         int
}

func (ts *TransactionSummary) CalculateBalance() {
	ts.Balance = ts.TotalIncome - ts.TotalExpenses
}

type BurnRateSummary struct {
	RecurringExpenses   float64
	RecurringCount      int
	OneTimeExpenses     float64
	OneTimeCount        int
	TotalBurn           float64
	ProjectedMonthly    float64
	ProjectedYearly     float64
}