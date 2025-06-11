package service

import (
	"fmt"
	"time"

	"budget-tracker/internal/models"
	"budget-tracker/internal/repository"
)

type TransactionService struct {
	repo            *repository.TransactionRepository
	currencyService *CurrencyService
	recurringRepo   *repository.RecurringTransactionRepository
}

func NewTransactionService(repo *repository.TransactionRepository, currencyService *CurrencyService) *TransactionService {
	return &TransactionService{
		repo:            repo,
		currencyService: currencyService,
	}
}

func (s *TransactionService) SetRecurringRepo(recurringRepo *repository.RecurringTransactionRepository) {
	s.recurringRepo = recurringRepo
}

func (s *TransactionService) Create(tx *models.Transaction) error {
	if err := tx.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if tx.Currency != "USD" {
		amountUSD, err := s.currencyService.ConvertToUSD(tx.Amount, tx.Currency)
		if err != nil {
			return fmt.Errorf("failed to convert currency: %w", err)
		}
		tx.AmountUSD = amountUSD
	} else {
		tx.AmountUSD = tx.Amount
	}

	return s.repo.Create(tx)
}

func (s *TransactionService) Update(tx *models.Transaction) error {
	if err := tx.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if tx.Currency != "USD" {
		amountUSD, err := s.currencyService.ConvertToUSD(tx.Amount, tx.Currency)
		if err != nil {
			return fmt.Errorf("failed to convert currency: %w", err)
		}
		tx.AmountUSD = amountUSD
	} else {
		tx.AmountUSD = tx.Amount
	}

	return s.repo.Update(tx)
}

func (s *TransactionService) Delete(id uint) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("transaction not found: %w", err)
	}

	return s.repo.Delete(id)
}

func (s *TransactionService) GetByID(id uint) (*models.Transaction, error) {
	return s.repo.GetByID(id)
}

func (s *TransactionService) GetAll() ([]*models.Transaction, error) {
	return s.repo.GetAll()
}

func (s *TransactionService) GetByDateRange(start, end time.Time) ([]*models.Transaction, error) {
	return s.repo.GetByDateRange(start, end)
}

func (s *TransactionService) GetByFilter(filter *models.TransactionFilter) ([]*models.Transaction, error) {
	return s.repo.GetByFilter(filter)
}

func (s *TransactionService) GetCurrentMonthSummary() (*models.TransactionSummary, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 1, 0).Add(-time.Second)
	
	return s.repo.GetSummary(start, end)
}

func (s *TransactionService) GetMonthSummary(year int, month time.Month) (*models.TransactionSummary, error) {
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0).Add(-time.Second)
	
	return s.repo.GetSummary(start, end)
}

func (s *TransactionService) GetYearSummary(year int) (*models.TransactionSummary, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(1, 0, 0).Add(-time.Second)
	
	return s.repo.GetSummary(start, end)
}

func (s *TransactionService) GetCategorySummary(start, end time.Time) ([]*models.CategoryWithTotal, error) {
	return s.repo.GetCategorySummary(start, end)
}

func (s *TransactionService) GetCurrentMonthCategorySummary() ([]*models.CategoryWithTotal, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 1, 0).Add(-time.Second)
	
	return s.repo.GetCategorySummary(start, end)
}

func (s *TransactionService) GetRecentTransactions(limit int) ([]*models.Transaction, error) {
	return s.repo.GetRecentTransactions(limit)
}

func (s *TransactionService) ImportTransactions(transactions []*models.Transaction) error {
	for _, tx := range transactions {
		if err := s.Create(tx); err != nil {
			return fmt.Errorf("failed to import transaction: %w", err)
		}
	}
	return nil
}

func (s *TransactionService) CountByCurrency(currency string) (int64, error) {
	return s.repo.CountByCurrency(currency)
}

func (s *TransactionService) GetCurrentMonthBurnRate() (*models.BurnRateSummary, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)
	
	// Get all expenses for the current month
	filter := models.TransactionFilter{
		Type:      models.TransactionTypeExpense,
		StartDate: startOfMonth,
		EndDate:   endOfMonth,
	}
	
	transactions, err := s.repo.GetByFilter(&filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	
	// Calculate burn rate
	burnRate := &models.BurnRateSummary{}
	
	for _, tx := range transactions {
		if tx.RecurringTransactionID != nil {
			burnRate.RecurringExpenses += tx.AmountUSD
			burnRate.RecurringCount++
		} else {
			burnRate.OneTimeExpenses += tx.AmountUSD
			burnRate.OneTimeCount++
		}
	}
	
	burnRate.TotalBurn = burnRate.RecurringExpenses + burnRate.OneTimeExpenses
	
	// Calculate projections based on active recurring transactions
	if s.recurringRepo != nil {
		activeRecurring, err := s.recurringRepo.GetActive()
		if err == nil {
			monthlyProjection := 0.0
			for _, recurring := range activeRecurring {
				if recurring.Type == models.TransactionTypeExpense {
					// Convert to monthly amount based on frequency
					monthlyAmount := s.calculateMonthlyAmount(recurring)
					monthlyProjection += monthlyAmount
				}
			}
			burnRate.ProjectedMonthly = monthlyProjection
			burnRate.ProjectedYearly = monthlyProjection * 12
		}
	} else {
		// Fallback to current month's recurring if repo not available
		burnRate.ProjectedMonthly = burnRate.RecurringExpenses
		burnRate.ProjectedYearly = burnRate.RecurringExpenses * 12
	}
	
	return burnRate, nil
}

func (s *TransactionService) calculateMonthlyAmount(recurring *models.RecurringTransaction) float64 {
	amount := recurring.Amount
	if recurring.Currency != "USD" {
		// Convert to USD if needed
		if amountUSD, err := s.currencyService.ConvertToUSD(amount, recurring.Currency); err == nil {
			amount = amountUSD
		}
	}
	
	// Convert to monthly based on frequency
	switch recurring.Frequency {
	case models.FrequencyDaily:
		return amount * 30.44 / float64(recurring.FrequencyValue) // Average days per month
	case models.FrequencyWeekly:
		return amount * 4.33 / float64(recurring.FrequencyValue) // Average weeks per month
	case models.FrequencyMonthly:
		return amount / float64(recurring.FrequencyValue)
	case models.FrequencyYearly:
		return amount / (12 * float64(recurring.FrequencyValue))
	default:
		return amount
	}
}