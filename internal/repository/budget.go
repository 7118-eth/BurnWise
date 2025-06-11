package repository

import (
	"time"

	"gorm.io/gorm"

	"burnwise/internal/models"
)

type BudgetRepository struct {
	db *gorm.DB
}

func NewBudgetRepository(db *gorm.DB) *BudgetRepository {
	return &BudgetRepository{db: db}
}

func (r *BudgetRepository) Create(budget *models.Budget) error {
	return r.db.Create(budget).Error
}

func (r *BudgetRepository) GetByID(id uint) (*models.Budget, error) {
	var budget models.Budget
	err := r.db.Preload("Category").First(&budget, id).Error
	if err != nil {
		return nil, err
	}
	return &budget, nil
}

func (r *BudgetRepository) Update(budget *models.Budget) error {
	return r.db.Save(budget).Error
}

func (r *BudgetRepository) Delete(id uint) error {
	return r.db.Delete(&models.Budget{}, id).Error
}

func (r *BudgetRepository) GetAll() ([]*models.Budget, error) {
	var budgets []*models.Budget
	err := r.db.Preload("Category").Find(&budgets).Error
	return budgets, err
}

func (r *BudgetRepository) GetActive() ([]*models.Budget, error) {
	now := time.Now()
	var budgets []*models.Budget
	
	err := r.db.Preload("Category").
		Where("start_date <= ?", now).
		Where("end_date IS NULL OR end_date >= ?", now).
		Find(&budgets).Error
	
	return budgets, err
}

func (r *BudgetRepository) GetByCategory(categoryID uint) ([]*models.Budget, error) {
	var budgets []*models.Budget
	err := r.db.Preload("Category").
		Where("category_id = ?", categoryID).
		Order("start_date DESC").
		Find(&budgets).Error
	return budgets, err
}

func (r *BudgetRepository) GetActiveByCategoryAndPeriod(categoryID uint, period models.BudgetPeriod) (*models.Budget, error) {
	now := time.Now()
	var budget models.Budget
	
	err := r.db.Preload("Category").
		Where("category_id = ? AND period = ?", categoryID, period).
		Where("start_date <= ?", now).
		Where("end_date IS NULL OR end_date >= ?", now).
		First(&budget).Error
	
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	
	if err != nil {
		return nil, err
	}
	
	return &budget, nil
}

func (r *BudgetRepository) GetByFilter(filter *models.BudgetFilter) ([]*models.Budget, error) {
	query := r.db.Preload("Category")

	if filter.CategoryID != 0 {
		query = query.Where("category_id = ?", filter.CategoryID)
	}

	if filter.Period != "" {
		query = query.Where("period = ?", filter.Period)
	}

	if filter.Active {
		now := time.Now()
		query = query.Where("start_date <= ?", now).
			Where("end_date IS NULL OR end_date >= ?", now)
	}

	if !filter.StartDate.IsZero() {
		query = query.Where("start_date >= ?", filter.StartDate)
	}

	if !filter.EndDate.IsZero() {
		query = query.Where("start_date <= ?", filter.EndDate)
	}

	var budgets []*models.Budget
	err := query.Order("start_date DESC").Find(&budgets).Error
	return budgets, err
}

func (r *BudgetRepository) GetSpentAmount(budgetID uint, start, end time.Time) (float64, error) {
	var budget models.Budget
	if err := r.db.First(&budget, budgetID).Error; err != nil {
		return 0, err
	}

	var spent float64
	err := r.db.Model(&models.Transaction{}).
		Select("COALESCE(SUM(amount_usd), 0)").
		Where("category_id = ? AND type = ? AND date >= ? AND date <= ?", 
			budget.CategoryID, 
			models.TransactionTypeExpense,
			start, 
			end).
		Scan(&spent).Error

	return spent, err
}

func (r *BudgetRepository) GetAllWithStatus() ([]*models.BudgetStatus, error) {
	budgets, err := r.GetActive()
	if err != nil {
		return nil, err
	}

	var statuses []*models.BudgetStatus
	for _, budget := range budgets {
		status := &models.BudgetStatus{
			Budget: *budget,
		}

		periodStart := budget.GetCurrentPeriodStart()
		periodEnd := budget.GetCurrentPeriodEnd()

		spent, err := r.GetSpentAmount(budget.ID, periodStart, periodEnd)
		if err != nil {
			return nil, err
		}

		status.Spent = spent
		status.Calculate()
		statuses = append(statuses, status)
	}

	return statuses, nil
}