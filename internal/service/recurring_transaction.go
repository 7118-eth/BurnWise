package service

import (
	"fmt"
	"time"

	"budget-tracker/internal/models"
	"budget-tracker/internal/repository"
)

type RecurringTransactionService struct {
	repo            *repository.RecurringTransactionRepository
	transactionRepo *repository.TransactionRepository
	currencyService *CurrencyService
}

func NewRecurringTransactionService(
	repo *repository.RecurringTransactionRepository,
	transactionRepo *repository.TransactionRepository,
	currencyService *CurrencyService,
) *RecurringTransactionService {
	return &RecurringTransactionService{
		repo:            repo,
		transactionRepo: transactionRepo,
		currencyService: currencyService,
	}
}

// Create creates a new recurring transaction
func (s *RecurringTransactionService) Create(rt *models.RecurringTransaction) error {
	if err := rt.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set initial next due date if not set
	if rt.NextDueDate.IsZero() {
		rt.NextDueDate = rt.StartDate
	}

	return s.repo.Create(rt)
}

// Update updates a recurring transaction
func (s *RecurringTransactionService) Update(rt *models.RecurringTransaction) error {
	if err := rt.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	existing, err := s.repo.GetByID(rt.ID)
	if err != nil {
		return fmt.Errorf("recurring transaction not found: %w", err)
	}

	// If frequency changed, recalculate next due date
	if existing.Frequency != rt.Frequency || existing.FrequencyValue != rt.FrequencyValue {
		if rt.LastProcessed != nil {
			rt.NextDueDate = rt.CalculateNextDueDate(*rt.LastProcessed)
		} else {
			rt.NextDueDate = rt.CalculateNextDueDate(rt.StartDate)
		}
	}

	return s.repo.Update(rt)
}

// Delete deletes a recurring transaction
func (s *RecurringTransactionService) Delete(id uint) error {
	// Check if any transactions have been generated
	count, err := s.repo.CountGeneratedTransactions(id)
	if err != nil {
		return err
	}

	if count > 0 {
		// Deactivate instead of delete if transactions exist
		return s.repo.Deactivate(id)
	}

	return s.repo.Delete(id)
}

// GetByID retrieves a recurring transaction by ID
func (s *RecurringTransactionService) GetByID(id uint) (*models.RecurringTransaction, error) {
	return s.repo.GetByID(id)
}

// GetAll retrieves all recurring transactions
func (s *RecurringTransactionService) GetAll() ([]*models.RecurringTransaction, error) {
	return s.repo.GetAll()
}

// GetActive retrieves all active recurring transactions
func (s *RecurringTransactionService) GetActive() ([]*models.RecurringTransaction, error) {
	return s.repo.GetActive()
}

// GetDue retrieves all recurring transactions due by a specific date
func (s *RecurringTransactionService) GetDue(asOf time.Time) ([]*models.RecurringTransaction, error) {
	return s.repo.GetDue(asOf)
}

// ProcessDueTransactions processes all due recurring transactions
func (s *RecurringTransactionService) ProcessDueTransactions(asOf time.Time) (int, error) {
	dueTransactions, err := s.repo.GetDue(asOf)
	if err != nil {
		return 0, fmt.Errorf("failed to get due transactions: %w", err)
	}

	processed := 0
	for _, rt := range dueTransactions {
		// Process all due dates up to asOf
		for rt.IsDue(asOf) {
			if err := s.processRecurringTransaction(rt, rt.NextDueDate); err != nil {
				// Log error but continue processing others
				fmt.Printf("Error processing recurring transaction %d: %v\n", rt.ID, err)
				break
			}
			processed++

			// Update next due date
			rt.NextDueDate = rt.CalculateNextDueDate(rt.NextDueDate)
			
			// Check if we should deactivate
			if rt.ShouldDeactivate(asOf) {
				rt.IsActive = false
				break
			}
		}

		// Update the recurring transaction
		if err := s.repo.Update(rt); err != nil {
			fmt.Printf("Error updating recurring transaction %d: %v\n", rt.ID, err)
		}
	}

	return processed, nil
}

// processRecurringTransaction processes a single occurrence of a recurring transaction
func (s *RecurringTransactionService) processRecurringTransaction(rt *models.RecurringTransaction, dueDate time.Time) error {
	// Check if this occurrence has been modified or skipped
	occurrence, err := s.repo.GetOccurrence(rt.ID, dueDate)
	if err != nil {
		return err
	}

	if occurrence != nil && occurrence.Action == models.OccurrenceActionSkip {
		// Skip this occurrence
		return nil
	}

	// Generate transaction
	tx := rt.GenerateTransaction(dueDate)

	// Apply any modifications from occurrence
	if occurrence != nil && occurrence.Action == models.OccurrenceActionModify {
		if occurrence.ModifiedAmount != nil {
			tx.Amount = *occurrence.ModifiedAmount
		}
		if occurrence.ModifiedDescription != nil {
			tx.Description = *occurrence.ModifiedDescription
		}
	}

	// Convert to USD
	amountUSD, err := s.currencyService.ConvertToUSD(tx.Amount, tx.Currency)
	if err != nil {
		return fmt.Errorf("failed to convert currency: %w", err)
	}
	tx.AmountUSD = amountUSD

	// Create the transaction
	if err := s.transactionRepo.Create(tx); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	// Update last processed date
	now := time.Now()
	rt.LastProcessed = &now

	return nil
}

// SkipOccurrence skips a specific occurrence of a recurring transaction
func (s *RecurringTransactionService) SkipOccurrence(recurringTransactionID uint, date time.Time, reason string) error {
	occurrence := &models.RecurringTransactionOccurrence{
		RecurringTransactionID: recurringTransactionID,
		OccurrenceDate:         date,
		Action:                 models.OccurrenceActionSkip,
		SkipReason:             &reason,
	}

	return s.repo.CreateOccurrence(occurrence)
}

// ModifyOccurrence modifies a specific occurrence of a recurring transaction
func (s *RecurringTransactionService) ModifyOccurrence(
	recurringTransactionID uint,
	date time.Time,
	amount *float64,
	description *string,
) error {
	occurrence := &models.RecurringTransactionOccurrence{
		RecurringTransactionID: recurringTransactionID,
		OccurrenceDate:         date,
		Action:                 models.OccurrenceActionModify,
		ModifiedAmount:         amount,
		ModifiedDescription:    description,
	}

	return s.repo.CreateOccurrence(occurrence)
}

// Pause pauses a recurring transaction
func (s *RecurringTransactionService) Pause(id uint) error {
	return s.repo.Deactivate(id)
}

// Resume resumes a recurring transaction
func (s *RecurringTransactionService) Resume(id uint) error {
	rt, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Update next due date to today or later
	now := time.Now()
	if rt.NextDueDate.Before(now) {
		// Calculate next due date from today
		rt.NextDueDate = rt.StartDate
		for rt.NextDueDate.Before(now) {
			rt.NextDueDate = rt.CalculateNextDueDate(rt.NextDueDate)
		}
		if err := s.repo.UpdateNextDueDate(id, rt.NextDueDate); err != nil {
			return err
		}
	}

	return s.repo.Activate(id)
}

// GetGeneratedTransactions retrieves all transactions generated from a recurring transaction
func (s *RecurringTransactionService) GetGeneratedTransactions(recurringTransactionID uint) ([]*models.Transaction, error) {
	return s.repo.GetGeneratedTransactions(recurringTransactionID)
}

// GetUpcoming retrieves upcoming occurrences for the next n days
func (s *RecurringTransactionService) GetUpcoming(days int) ([]*models.RecurringTransaction, error) {
	endDate := time.Now().AddDate(0, 0, days)
	
	active, err := s.repo.GetActive()
	if err != nil {
		return nil, err
	}

	var upcoming []*models.RecurringTransaction
	for _, rt := range active {
		if rt.NextDueDate.Before(endDate) {
			upcoming = append(upcoming, rt)
		}
	}

	return upcoming, nil
}

// GetExpiring retrieves recurring transactions expiring soon
func (s *RecurringTransactionService) GetExpiring(days int) ([]*models.RecurringTransaction, error) {
	start := time.Now()
	end := start.AddDate(0, 0, days)
	return s.repo.GetExpiring(start, end)
}

// CalculateProjectedAmount calculates the projected amount for a period
func (s *RecurringTransactionService) CalculateProjectedAmount(startDate, endDate time.Time) (float64, error) {
	active, err := s.repo.GetActive()
	if err != nil {
		return 0, err
	}

	totalUSD := 0.0
	for _, rt := range active {
		// Skip if starts after end date
		if rt.StartDate.After(endDate) {
			continue
		}

		// Calculate occurrences in the period
		occurrences := 0
		currentDate := rt.NextDueDate
		
		// If next due date is before start, advance to start
		for currentDate.Before(startDate) {
			currentDate = rt.CalculateNextDueDate(currentDate)
		}

		// Count occurrences within the period
		for !currentDate.After(endDate) {
			if rt.EndDate == nil || !currentDate.After(*rt.EndDate) {
				occurrences++
			}
			currentDate = rt.CalculateNextDueDate(currentDate)
		}

		if occurrences > 0 {
			// Convert to USD for aggregation
			amountUSD, err := s.currencyService.ConvertToUSD(rt.Amount, rt.Currency)
			if err != nil {
				return 0, fmt.Errorf("failed to convert currency: %w", err)
			}

			projectedAmount := amountUSD * float64(occurrences)
			if rt.Type == models.TransactionTypeIncome {
				totalUSD += projectedAmount
			} else {
				totalUSD -= projectedAmount
			}
		}
	}

	return totalUSD, nil
}