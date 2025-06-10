package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"budget-tracker/internal/models"
)

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(tx *models.Transaction) error {
	return r.db.Create(tx).Error
}

func (r *TransactionRepository) GetByID(id uint) (*models.Transaction, error) {
	var tx models.Transaction
	err := r.db.Preload("Category").First(&tx, id).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *TransactionRepository) Update(tx *models.Transaction) error {
	return r.db.Save(tx).Error
}

func (r *TransactionRepository) Delete(id uint) error {
	return r.db.Delete(&models.Transaction{}, id).Error
}

func (r *TransactionRepository) GetAll() ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	err := r.db.Preload("Category").Order("date DESC").Find(&transactions).Error
	return transactions, err
}

func (r *TransactionRepository) GetByDateRange(start, end time.Time) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	err := r.db.Preload("Category").
		Where("date >= ? AND date <= ?", start, end).
		Order("date DESC").
		Find(&transactions).Error
	return transactions, err
}

func (r *TransactionRepository) GetByCategory(categoryID uint) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	err := r.db.Preload("Category").
		Where("category_id = ?", categoryID).
		Order("date DESC").
		Find(&transactions).Error
	return transactions, err
}

func (r *TransactionRepository) GetByFilter(filter *models.TransactionFilter) ([]*models.Transaction, error) {
	query := r.db.Preload("Category")

	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}

	if filter.CategoryID != 0 {
		query = query.Where("category_id = ?", filter.CategoryID)
	}

	if !filter.StartDate.IsZero() {
		query = query.Where("date >= ?", filter.StartDate)
	}

	if !filter.EndDate.IsZero() {
		query = query.Where("date <= ?", filter.EndDate)
	}

	if filter.MinAmount > 0 {
		query = query.Where("amount_usd >= ?", filter.MinAmount)
	}

	if filter.MaxAmount > 0 {
		query = query.Where("amount_usd <= ?", filter.MaxAmount)
	}

	if filter.Currency != "" {
		query = query.Where("currency = ?", filter.Currency)
	}

	if filter.Search != "" {
		searchPattern := fmt.Sprintf("%%%s%%", filter.Search)
		query = query.Where("description LIKE ?", searchPattern)
	}

	var transactions []*models.Transaction
	err := query.Order("date DESC").Find(&transactions).Error
	return transactions, err
}

func (r *TransactionRepository) GetSummary(start, end time.Time) (*models.TransactionSummary, error) {
	summary := &models.TransactionSummary{}

	var incomeResult struct {
		Total float64
		Count int
	}
	r.db.Model(&models.Transaction{}).
		Select("SUM(amount_usd) as total, COUNT(*) as count").
		Where("type = ? AND date >= ? AND date <= ?", models.TransactionTypeIncome, start, end).
		Scan(&incomeResult)

	var expenseResult struct {
		Total float64
		Count int
	}
	r.db.Model(&models.Transaction{}).
		Select("SUM(amount_usd) as total, COUNT(*) as count").
		Where("type = ? AND date >= ? AND date <= ?", models.TransactionTypeExpense, start, end).
		Scan(&expenseResult)

	summary.TotalIncome = incomeResult.Total
	summary.TotalExpenses = expenseResult.Total
	summary.Count = incomeResult.Count + expenseResult.Count
	summary.CalculateBalance()

	return summary, nil
}

func (r *TransactionRepository) GetCategorySummary(start, end time.Time) ([]*models.CategoryWithTotal, error) {
	var results []*models.CategoryWithTotal

	err := r.db.Table("transactions").
		Select("categories.*, SUM(transactions.amount_usd) as total, COUNT(transactions.id) as count").
		Joins("JOIN categories ON categories.id = transactions.category_id").
		Where("transactions.date >= ? AND transactions.date <= ?", start, end).
		Where("transactions.deleted_at IS NULL").
		Group("categories.id").
		Order("total DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	var totalAmount float64
	for _, result := range results {
		totalAmount += result.Total
	}

	for _, result := range results {
		if totalAmount > 0 {
			result.Percentage = (result.Total / totalAmount) * 100
		}
	}

	return results, nil
}

func (r *TransactionRepository) GetRecentTransactions(limit int) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	err := r.db.Preload("Category").
		Order("date DESC").
		Limit(limit).
		Find(&transactions).Error
	return transactions, err
}