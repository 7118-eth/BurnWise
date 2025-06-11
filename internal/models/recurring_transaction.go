package models

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type RecurrenceFrequency string

const (
	FrequencyDaily   RecurrenceFrequency = "daily"
	FrequencyWeekly  RecurrenceFrequency = "weekly"
	FrequencyMonthly RecurrenceFrequency = "monthly"
	FrequencyYearly  RecurrenceFrequency = "yearly"
)

type RecurringTransaction struct {
	ID             uint                `gorm:"primaryKey" json:"id"`
	Type           TransactionType     `gorm:"type:varchar(20);not null" json:"type"`
	Amount         float64             `gorm:"not null" json:"amount"`
	Currency       string              `gorm:"type:varchar(3);not null;default:'USD'" json:"currency"`
	CategoryID     uint                `gorm:"not null" json:"category_id"`
	Description    string              `gorm:"type:varchar(500)" json:"description"`
	Frequency      RecurrenceFrequency `gorm:"type:varchar(20);not null" json:"frequency"`
	FrequencyValue int                 `gorm:"default:1" json:"frequency_value"` // e.g., every 2 weeks
	StartDate      time.Time           `gorm:"not null" json:"start_date"`
	EndDate        *time.Time          `json:"end_date,omitempty"`
	LastProcessed  *time.Time          `json:"last_processed,omitempty"`
	NextDueDate    time.Time           `gorm:"not null" json:"next_due_date"`
	IsActive       bool                `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
	DeletedAt      gorm.DeletedAt      `gorm:"index" json:"deleted_at,omitempty"`

	// Relationships
	Category     Category      `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Transactions []Transaction `gorm:"foreignKey:RecurringTransactionID" json:"transactions,omitempty"`
}

// RecurringTransactionOccurrence tracks modifications to specific occurrences
type RecurringTransactionOccurrence struct {
	ID                     uint            `gorm:"primaryKey" json:"id"`
	RecurringTransactionID uint            `gorm:"not null" json:"recurring_transaction_id"`
	OccurrenceDate         time.Time       `gorm:"not null" json:"occurrence_date"`
	Action                 string          `gorm:"type:varchar(20);not null" json:"action"` // skip, modify
	ModifiedAmount         *float64        `json:"modified_amount,omitempty"`
	ModifiedDescription    *string         `json:"modified_description,omitempty"`
	SkipReason             *string         `json:"skip_reason,omitempty"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`
	DeletedAt              gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	RecurringTransaction RecurringTransaction `gorm:"foreignKey:RecurringTransactionID" json:"recurring_transaction,omitempty"`
}

const (
	OccurrenceActionSkip   = "skip"
	OccurrenceActionModify = "modify"
)

func (rt *RecurringTransaction) Validate() error {
	if rt.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}

	if rt.Currency == "" {
		rt.Currency = "USD"
	}

	if rt.Type != TransactionTypeIncome && rt.Type != TransactionTypeExpense {
		return errors.New("invalid transaction type")
	}

	if rt.CategoryID == 0 {
		return errors.New("category is required")
	}

	if rt.FrequencyValue < 1 {
		rt.FrequencyValue = 1
	}

	// Validate frequency
	switch rt.Frequency {
	case FrequencyDaily, FrequencyWeekly, FrequencyMonthly, FrequencyYearly:
		// Valid
	default:
		return fmt.Errorf("invalid frequency: %s", rt.Frequency)
	}

	// Ensure start date is set
	if rt.StartDate.IsZero() {
		rt.StartDate = time.Now()
	}

	// Validate end date is after start date if set
	if rt.EndDate != nil && !rt.EndDate.After(rt.StartDate) {
		return errors.New("end date must be after start date")
	}

	// Calculate initial next due date if not set
	if rt.NextDueDate.IsZero() {
		rt.NextDueDate = rt.StartDate
	}

	return nil
}

func (rt *RecurringTransaction) BeforeCreate(tx *gorm.DB) error {
	return rt.Validate()
}

func (rt *RecurringTransaction) BeforeUpdate(tx *gorm.DB) error {
	return rt.Validate()
}

// CalculateNextDueDate calculates the next due date based on frequency
func (rt *RecurringTransaction) CalculateNextDueDate(from time.Time) time.Time {
	switch rt.Frequency {
	case FrequencyDaily:
		return from.AddDate(0, 0, rt.FrequencyValue)
	case FrequencyWeekly:
		return from.AddDate(0, 0, rt.FrequencyValue*7)
	case FrequencyMonthly:
		return from.AddDate(0, rt.FrequencyValue, 0)
	case FrequencyYearly:
		return from.AddDate(rt.FrequencyValue, 0, 0)
	default:
		return from
	}
}

// IsDue checks if the recurring transaction is due for processing
func (rt *RecurringTransaction) IsDue(asOf time.Time) bool {
	if !rt.IsActive {
		return false
	}

	// Check if past end date
	if rt.EndDate != nil && asOf.After(*rt.EndDate) {
		return false
	}

	// Check if due
	return !rt.NextDueDate.After(asOf)
}

// ShouldDeactivate checks if the recurring transaction should be deactivated
func (rt *RecurringTransaction) ShouldDeactivate(asOf time.Time) bool {
	if rt.EndDate == nil {
		return false
	}
	return asOf.After(*rt.EndDate)
}

// GenerateTransaction creates a transaction from this recurring transaction
func (rt *RecurringTransaction) GenerateTransaction(date time.Time) *Transaction {
	return &Transaction{
		Type:                   rt.Type,
		Amount:                 rt.Amount,
		Currency:               rt.Currency,
		CategoryID:             rt.CategoryID,
		Description:            rt.Description,
		Date:                   date,
		RecurringTransactionID: &rt.ID,
	}
}

// GetFrequencyDisplay returns a human-readable frequency description
func (rt *RecurringTransaction) GetFrequencyDisplay() string {
	if rt.FrequencyValue == 1 {
		switch rt.Frequency {
		case FrequencyDaily:
			return "Daily"
		case FrequencyWeekly:
			return "Weekly"
		case FrequencyMonthly:
			return "Monthly"
		case FrequencyYearly:
			return "Yearly"
		}
	}

	unit := ""
	switch rt.Frequency {
	case FrequencyDaily:
		unit = "days"
	case FrequencyWeekly:
		unit = "weeks"
	case FrequencyMonthly:
		unit = "months"
	case FrequencyYearly:
		unit = "years"
	}

	return fmt.Sprintf("Every %d %s", rt.FrequencyValue, unit)
}

// IsValidFrequency checks if a frequency string is valid
func IsValidFrequency(freq string) bool {
	switch RecurrenceFrequency(freq) {
	case FrequencyDaily, FrequencyWeekly, FrequencyMonthly, FrequencyYearly:
		return true
	default:
		return false
	}
}

// GetAllFrequencies returns all available frequencies
func GetAllFrequencies() []RecurrenceFrequency {
	return []RecurrenceFrequency{
		FrequencyDaily,
		FrequencyWeekly,
		FrequencyMonthly,
		FrequencyYearly,
	}
}