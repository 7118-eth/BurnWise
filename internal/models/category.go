package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	Name      string          `gorm:"type:varchar(100);not null;uniqueIndex:idx_category_name_type" json:"name"`
	Type      TransactionType `gorm:"type:varchar(20);not null;uniqueIndex:idx_category_name_type" json:"type"`
	Icon      string          `gorm:"type:varchar(10)" json:"icon"`
	Color     string          `gorm:"type:varchar(7)" json:"color"`
	ParentID  *uint           `json:"parent_id,omitempty"`
	IsDefault bool            `gorm:"default:false" json:"is_default"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt gorm.DeletedAt  `gorm:"index" json:"deleted_at,omitempty"`

	Parent       *Category     `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Transactions []Transaction `gorm:"foreignKey:CategoryID" json:"transactions,omitempty"`
}

func (c *Category) Validate() error {
	if c.Name == "" {
		return errors.New("category name is required")
	}

	if c.Type != TransactionTypeIncome && c.Type != TransactionTypeExpense {
		return errors.New("invalid category type")
	}

	if c.Color != "" && len(c.Color) != 7 {
		return errors.New("color must be a hex code (e.g., #FF5733)")
	}

	return nil
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	return c.Validate()
}

var DefaultIncomeCategories = []Category{
	{Name: "Salary", Icon: "💼", Type: TransactionTypeIncome, IsDefault: true, Color: "#4CAF50"},
	{Name: "Freelance", Icon: "💻", Type: TransactionTypeIncome, IsDefault: true, Color: "#2196F3"},
	{Name: "Investments", Icon: "📈", Type: TransactionTypeIncome, IsDefault: true, Color: "#00BCD4"},
	{Name: "Other Income", Icon: "💰", Type: TransactionTypeIncome, IsDefault: true, Color: "#009688"},
}

var DefaultExpenseCategories = []Category{
	// Housing & Living
	{Name: "Housing", Icon: "🏠", Type: TransactionTypeExpense, IsDefault: true, Color: "#FF5722"},
	{Name: "Utilities", Icon: "💡", Type: TransactionTypeExpense, IsDefault: true, Color: "#FF9800"},
	{Name: "Living", Icon: "🛒", Type: TransactionTypeExpense, IsDefault: true, Color: "#FFC107"},
	{Name: "Transportation", Icon: "🚗", Type: TransactionTypeExpense, IsDefault: true, Color: "#FFEB3B"},
	
	// Business & Technology
	{Name: "Technology", Icon: "💻", Type: TransactionTypeExpense, IsDefault: true, Color: "#2196F3"},
	{Name: "AI Tools", Icon: "🤖", Type: TransactionTypeExpense, IsDefault: true, Color: "#9C27B0"},
	{Name: "Cloud Services", Icon: "☁️", Type: TransactionTypeExpense, IsDefault: true, Color: "#3F51B5"},
	{Name: "Business", Icon: "💼", Type: TransactionTypeExpense, IsDefault: true, Color: "#00BCD4"},
	
	// Personal
	{Name: "Healthcare", Icon: "💊", Type: TransactionTypeExpense, IsDefault: true, Color: "#4CAF50"},
	{Name: "Personal", Icon: "👤", Type: TransactionTypeExpense, IsDefault: true, Color: "#009688"},
	{Name: "Other", Icon: "💸", Type: TransactionTypeExpense, IsDefault: true, Color: "#607D8B"},
}

func GetDefaultCategories() []Category {
	categories := make([]Category, 0, len(DefaultIncomeCategories)+len(DefaultExpenseCategories))
	categories = append(categories, DefaultIncomeCategories...)
	categories = append(categories, DefaultExpenseCategories...)
	return categories
}

type CategoryWithTotal struct {
	Category
	Total     float64 `json:"total"`
	Count     int     `json:"count"`
	Percentage float64 `json:"percentage"`
}