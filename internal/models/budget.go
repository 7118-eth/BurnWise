package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type BudgetPeriod string

const (
	BudgetPeriodMonthly BudgetPeriod = "monthly"
	BudgetPeriodYearly  BudgetPeriod = "yearly"
)

type Budget struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Name       string         `gorm:"type:varchar(100);not null" json:"name"`
	CategoryID uint           `gorm:"not null" json:"category_id"`
	Amount     float64        `gorm:"not null" json:"amount"`
	Period     BudgetPeriod   `gorm:"type:varchar(20);not null" json:"period"`
	StartDate  time.Time      `gorm:"not null" json:"start_date"`
	EndDate    *time.Time     `json:"end_date,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Category Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

func (b *Budget) Validate() error {
	if b.Name == "" {
		return errors.New("budget name is required")
	}

	if b.CategoryID == 0 {
		return errors.New("category is required")
	}

	if b.Amount <= 0 {
		return errors.New("budget amount must be positive")
	}

	if b.Period != BudgetPeriodMonthly && b.Period != BudgetPeriodYearly {
		return errors.New("invalid budget period")
	}

	if b.StartDate.IsZero() {
		return errors.New("start date is required")
	}

	if b.EndDate != nil && b.EndDate.Before(b.StartDate) {
		return errors.New("end date must be after start date")
	}

	return nil
}

func (b *Budget) BeforeCreate(tx *gorm.DB) error {
	return b.Validate()
}

func (b *Budget) IsActive() bool {
	now := time.Now()
	return now.After(b.StartDate) && (b.EndDate == nil || now.Before(*b.EndDate))
}

func (b *Budget) GetCurrentPeriodStart() time.Time {
	now := time.Now()
	
	switch b.Period {
	case BudgetPeriodMonthly:
		year, month, _ := now.Date()
		return time.Date(year, month, 1, 0, 0, 0, 0, now.Location())
	case BudgetPeriodYearly:
		year := now.Year()
		return time.Date(year, 1, 1, 0, 0, 0, 0, now.Location())
	default:
		return b.StartDate
	}
}

func (b *Budget) GetCurrentPeriodEnd() time.Time {
	start := b.GetCurrentPeriodStart()
	
	switch b.Period {
	case BudgetPeriodMonthly:
		return start.AddDate(0, 1, 0).Add(-time.Second)
	case BudgetPeriodYearly:
		return start.AddDate(1, 0, 0).Add(-time.Second)
	default:
		if b.EndDate != nil {
			return *b.EndDate
		}
		return time.Now().AddDate(10, 0, 0) // Far future
	}
}

type BudgetStatus struct {
	Budget       Budget  `json:"budget"`
	Spent        float64 `json:"spent"`
	Remaining    float64 `json:"remaining"`
	PercentUsed  float64 `json:"percent_used"`
	IsOverBudget bool    `json:"is_over_budget"`
	DaysLeft     int     `json:"days_left"`
	DailyBudget  float64 `json:"daily_budget"`
}

func (bs *BudgetStatus) Calculate() {
	bs.Remaining = bs.Budget.Amount - bs.Spent
	bs.PercentUsed = (bs.Spent / bs.Budget.Amount) * 100
	bs.IsOverBudget = bs.Spent > bs.Budget.Amount
	
	end := bs.Budget.GetCurrentPeriodEnd()
	now := time.Now()
	if end.After(now) {
		bs.DaysLeft = int(end.Sub(now).Hours() / 24) + 1
		bs.DailyBudget = bs.Remaining / float64(bs.DaysLeft)
		if bs.DailyBudget < 0 {
			bs.DailyBudget = 0
		}
	}
}

type BudgetFilter struct {
	CategoryID uint
	Period     BudgetPeriod
	Active     bool
	StartDate  time.Time
	EndDate    time.Time
}