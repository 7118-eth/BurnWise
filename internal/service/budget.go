package service

import (
	"fmt"
	"time"

	"burnwise/internal/models"
	"burnwise/internal/repository"
)

type BudgetService struct {
	budgetRepo *repository.BudgetRepository
	txRepo     *repository.TransactionRepository
}

func NewBudgetService(budgetRepo *repository.BudgetRepository, txRepo *repository.TransactionRepository) *BudgetService {
	return &BudgetService{
		budgetRepo: budgetRepo,
		txRepo:     txRepo,
	}
}

func (s *BudgetService) Create(budget *models.Budget) error {
	if err := budget.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	existing, err := s.budgetRepo.GetActiveByCategoryAndPeriod(budget.CategoryID, budget.Period)
	if err != nil {
		return err
	}

	if existing != nil {
		return fmt.Errorf("active budget already exists for this category and period")
	}

	return s.budgetRepo.Create(budget)
}

func (s *BudgetService) Update(budget *models.Budget) error {
	if err := budget.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	existing, err := s.budgetRepo.GetActiveByCategoryAndPeriod(budget.CategoryID, budget.Period)
	if err != nil {
		return err
	}

	if existing != nil && existing.ID != budget.ID {
		return fmt.Errorf("another active budget exists for this category and period")
	}

	return s.budgetRepo.Update(budget)
}

func (s *BudgetService) Delete(id uint) error {
	_, err := s.budgetRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("budget not found: %w", err)
	}

	return s.budgetRepo.Delete(id)
}

func (s *BudgetService) GetByID(id uint) (*models.Budget, error) {
	return s.budgetRepo.GetByID(id)
}

func (s *BudgetService) GetAll() ([]*models.Budget, error) {
	return s.budgetRepo.GetAll()
}

func (s *BudgetService) GetActive() ([]*models.Budget, error) {
	return s.budgetRepo.GetActive()
}

func (s *BudgetService) GetStatus(budgetID uint) (*models.BudgetStatus, error) {
	budget, err := s.budgetRepo.GetByID(budgetID)
	if err != nil {
		return nil, err
	}

	periodStart := budget.GetCurrentPeriodStart()
	periodEnd := budget.GetCurrentPeriodEnd()

	spent, err := s.budgetRepo.GetSpentAmount(budgetID, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}

	status := &models.BudgetStatus{
		Budget: *budget,
		Spent:  spent,
	}
	status.Calculate()

	return status, nil
}

func (s *BudgetService) GetAllStatuses() ([]*models.BudgetStatus, error) {
	return s.budgetRepo.GetAllWithStatus()
}

func (s *BudgetService) CheckOverspending(budgetID uint) (bool, float64, error) {
	status, err := s.GetStatus(budgetID)
	if err != nil {
		return false, 0, err
	}

	if status.IsOverBudget {
		overspent := status.Spent - status.Budget.Amount
		return true, overspent, nil
	}

	return false, 0, nil
}

func (s *BudgetService) GetCategoryBudgetStatus(categoryID uint) (*models.BudgetStatus, error) {
	monthlyBudget, _ := s.budgetRepo.GetActiveByCategoryAndPeriod(categoryID, models.BudgetPeriodMonthly)
	yearlyBudget, _ := s.budgetRepo.GetActiveByCategoryAndPeriod(categoryID, models.BudgetPeriodYearly)

	if monthlyBudget != nil {
		return s.GetStatus(monthlyBudget.ID)
	}

	if yearlyBudget != nil {
		return s.GetStatus(yearlyBudget.ID)
	}

	return nil, fmt.Errorf("no active budget found for category")
}

func (s *BudgetService) GetBudgetProgress() (map[uint]*models.BudgetStatus, error) {
	statuses, err := s.GetAllStatuses()
	if err != nil {
		return nil, err
	}

	progressMap := make(map[uint]*models.BudgetStatus)
	for _, status := range statuses {
		progressMap[status.Budget.CategoryID] = status
	}

	return progressMap, nil
}

func (s *BudgetService) CreateMonthlyBudgets(budgets map[uint]float64) error {
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	for categoryID, amount := range budgets {
		budget := &models.Budget{
			Name:       fmt.Sprintf("Monthly Budget - %s %d", now.Month().String(), now.Year()),
			CategoryID: categoryID,
			Amount:     amount,
			Period:     models.BudgetPeriodMonthly,
			StartDate:  startDate,
		}

		if err := s.Create(budget); err != nil {
			return err
		}
	}

	return nil
}