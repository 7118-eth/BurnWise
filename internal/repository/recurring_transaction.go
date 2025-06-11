package repository

import (
	"time"

	"gorm.io/gorm"

	"burnwise/internal/models"
)

type RecurringTransactionRepository struct {
	db *gorm.DB
}

func NewRecurringTransactionRepository(db *gorm.DB) *RecurringTransactionRepository {
	return &RecurringTransactionRepository{db: db}
}

// Create creates a new recurring transaction
func (r *RecurringTransactionRepository) Create(rt *models.RecurringTransaction) error {
	return r.db.Create(rt).Error
}

// GetByID retrieves a recurring transaction by ID
func (r *RecurringTransactionRepository) GetByID(id uint) (*models.RecurringTransaction, error) {
	var rt models.RecurringTransaction
	err := r.db.Preload("Category").First(&rt, id).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

// Update updates a recurring transaction
func (r *RecurringTransactionRepository) Update(rt *models.RecurringTransaction) error {
	return r.db.Save(rt).Error
}

// Delete soft deletes a recurring transaction
func (r *RecurringTransactionRepository) Delete(id uint) error {
	return r.db.Delete(&models.RecurringTransaction{}, id).Error
}

// GetAll retrieves all recurring transactions
func (r *RecurringTransactionRepository) GetAll() ([]*models.RecurringTransaction, error) {
	var rts []*models.RecurringTransaction
	err := r.db.Preload("Category").Order("next_due_date ASC").Find(&rts).Error
	return rts, err
}

// GetActive retrieves all active recurring transactions
func (r *RecurringTransactionRepository) GetActive() ([]*models.RecurringTransaction, error) {
	var rts []*models.RecurringTransaction
	err := r.db.Preload("Category").
		Where("is_active = ?", true).
		Order("next_due_date ASC").
		Find(&rts).Error
	return rts, err
}

// GetDue retrieves all recurring transactions due by a specific date
func (r *RecurringTransactionRepository) GetDue(asOf time.Time) ([]*models.RecurringTransaction, error) {
	var rts []*models.RecurringTransaction
	err := r.db.Preload("Category").
		Where("is_active = ? AND next_due_date <= ?", true, asOf).
		Where("end_date IS NULL OR end_date >= ?", asOf).
		Order("next_due_date ASC").
		Find(&rts).Error
	return rts, err
}

// GetByCategory retrieves all recurring transactions for a specific category
func (r *RecurringTransactionRepository) GetByCategory(categoryID uint) ([]*models.RecurringTransaction, error) {
	var rts []*models.RecurringTransaction
	err := r.db.Preload("Category").
		Where("category_id = ?", categoryID).
		Order("next_due_date ASC").
		Find(&rts).Error
	return rts, err
}

// UpdateNextDueDate updates the next due date for a recurring transaction
func (r *RecurringTransactionRepository) UpdateNextDueDate(id uint, nextDueDate time.Time) error {
	// Use UpdateColumn to skip hooks
	return r.db.Model(&models.RecurringTransaction{}).
		Where("id = ?", id).
		UpdateColumn("next_due_date", nextDueDate).Error
}

// UpdateLastProcessed updates the last processed date for a recurring transaction
func (r *RecurringTransactionRepository) UpdateLastProcessed(id uint, lastProcessed time.Time) error {
	// Use UpdateColumns to skip hooks
	return r.db.Model(&models.RecurringTransaction{}).
		Where("id = ?", id).
		UpdateColumns(map[string]interface{}{
			"last_processed": lastProcessed,
			"next_due_date":  lastProcessed, // This will be recalculated by the service
		}).Error
}

// Deactivate deactivates a recurring transaction
func (r *RecurringTransactionRepository) Deactivate(id uint) error {
	// Use UpdateColumn to skip hooks
	return r.db.Model(&models.RecurringTransaction{}).
		Where("id = ?", id).
		UpdateColumn("is_active", false).Error
}

// Activate activates a recurring transaction
func (r *RecurringTransactionRepository) Activate(id uint) error {
	// Use UpdateColumn to skip hooks
	return r.db.Model(&models.RecurringTransaction{}).
		Where("id = ?", id).
		UpdateColumn("is_active", true).Error
}

// CreateOccurrence creates a recurring transaction occurrence record
func (r *RecurringTransactionRepository) CreateOccurrence(occurrence *models.RecurringTransactionOccurrence) error {
	return r.db.Create(occurrence).Error
}

// GetOccurrence retrieves an occurrence for a specific date
func (r *RecurringTransactionRepository) GetOccurrence(recurringTransactionID uint, date time.Time) (*models.RecurringTransactionOccurrence, error) {
	var occurrence models.RecurringTransactionOccurrence
	err := r.db.Where("recurring_transaction_id = ? AND DATE(occurrence_date) = DATE(?)", 
		recurringTransactionID, date).
		First(&occurrence).Error
	
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &occurrence, err
}

// GetOccurrences retrieves all occurrences for a recurring transaction
func (r *RecurringTransactionRepository) GetOccurrences(recurringTransactionID uint) ([]*models.RecurringTransactionOccurrence, error) {
	var occurrences []*models.RecurringTransactionOccurrence
	err := r.db.Where("recurring_transaction_id = ?", recurringTransactionID).
		Order("occurrence_date DESC").
		Find(&occurrences).Error
	return occurrences, err
}

// GetGeneratedTransactions retrieves all transactions generated from a recurring transaction
func (r *RecurringTransactionRepository) GetGeneratedTransactions(recurringTransactionID uint) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	err := r.db.Where("recurring_transaction_id = ?", recurringTransactionID).
		Order("date DESC").
		Find(&transactions).Error
	return transactions, err
}

// CountGeneratedTransactions counts transactions generated from a recurring transaction
func (r *RecurringTransactionRepository) CountGeneratedTransactions(recurringTransactionID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Transaction{}).
		Where("recurring_transaction_id = ?", recurringTransactionID).
		Count(&count).Error
	return count, err
}

// GetExpiring retrieves recurring transactions expiring within a date range
func (r *RecurringTransactionRepository) GetExpiring(start, end time.Time) ([]*models.RecurringTransaction, error) {
	var rts []*models.RecurringTransaction
	err := r.db.Preload("Category").
		Where("is_active = ? AND end_date IS NOT NULL AND end_date BETWEEN ? AND ?", 
			true, start, end).
		Order("end_date ASC").
		Find(&rts).Error
	return rts, err
}