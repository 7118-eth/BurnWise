package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"burnwise/internal/models"
	"burnwise/internal/repository"
	test "burnwise/test/helpers"
)

func TestTransactionService_Create(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := repository.NewTransactionRepository(db)
	
	tempDir := t.TempDir()
	settingsService, err := NewSettingsService(tempDir)
	require.NoError(t, err)
	currencyService := NewCurrencyService(settingsService)
	
	service := NewTransactionService(repo, currencyService)
	
	category := test.CreateTestCategory(t, db, "Food", models.TransactionTypeExpense)
	
	tx := &models.Transaction{
		Type:        models.TransactionTypeExpense,
		Amount:      50.00,
		Currency:    "USD",
		CategoryID:  category.ID,
		Description: "Test transaction",
		Date:        time.Now(),
	}
	
	err = service.Create(tx)
	require.NoError(t, err)
	assert.Greater(t, tx.ID, uint(0))
	assert.Equal(t, tx.Amount, tx.AmountUSD)
}

func TestTransactionService_CreateWithCurrency(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := repository.NewTransactionRepository(db)
	
	tempDir := t.TempDir()
	settingsService, err := NewSettingsService(tempDir)
	require.NoError(t, err)
	currencyService := NewCurrencyService(settingsService)
	
	service := NewTransactionService(repo, currencyService)
	
	category := test.CreateTestCategory(t, db, "Food", models.TransactionTypeExpense)
	
	tx := &models.Transaction{
		Type:        models.TransactionTypeExpense,
		Amount:      100.00,
		Currency:    "AED",
		CategoryID:  category.ID,
		Description: "Test transaction in AED",
		Date:        time.Now(),
	}
	
	err = service.Create(tx)
	require.NoError(t, err)
	assert.Greater(t, tx.ID, uint(0))
	assert.InDelta(t, 27.23, tx.AmountUSD, 0.01)
}

func TestTransactionService_GetCurrentMonthSummary(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := repository.NewTransactionRepository(db)
	
	tempDir := t.TempDir()
	settingsService, err := NewSettingsService(tempDir)
	require.NoError(t, err)
	currencyService := NewCurrencyService(settingsService)
	
	service := NewTransactionService(repo, currencyService)
	
	incomeCategory := test.CreateTestCategory(t, db, "Salary", models.TransactionTypeIncome)
	expenseCategory := test.CreateTestCategory(t, db, "Food", models.TransactionTypeExpense)
	
	// Create income transaction
	income := &models.Transaction{
		Type:        models.TransactionTypeIncome,
		Amount:      5000.00,
		Currency:    "USD",
		CategoryID:  incomeCategory.ID,
		Description: "Monthly salary",
		Date:        time.Now(),
	}
	require.NoError(t, service.Create(income))
	
	// Create expense transactions
	expense1 := &models.Transaction{
		Type:        models.TransactionTypeExpense,
		Amount:      100.00,
		Currency:    "USD",
		CategoryID:  expenseCategory.ID,
		Description: "Groceries",
		Date:        time.Now(),
	}
	require.NoError(t, service.Create(expense1))
	
	expense2 := &models.Transaction{
		Type:        models.TransactionTypeExpense,
		Amount:      50.00,
		Currency:    "USD",
		CategoryID:  expenseCategory.ID,
		Description: "Lunch",
		Date:        time.Now(),
	}
	require.NoError(t, service.Create(expense2))
	
	// Get summary
	summary, err := service.GetCurrentMonthSummary()
	require.NoError(t, err)
	
	assert.Equal(t, 5000.0, summary.TotalIncome)
	assert.Equal(t, 150.0, summary.TotalExpenses)
	assert.Equal(t, 4850.0, summary.Balance)
	assert.Equal(t, 3, summary.Count)
}

func TestTransactionService_GetCurrentMonthBurnRate(t *testing.T) {
	db := test.SetupTestDB(t)
	txRepo := repository.NewTransactionRepository(db)
	recurringRepo := repository.NewRecurringTransactionRepository(db)
	
	tempDir := t.TempDir()
	settingsService, err := NewSettingsService(tempDir)
	require.NoError(t, err)
	currencyService := NewCurrencyService(settingsService)
	
	service := NewTransactionService(txRepo, currencyService)
	service.SetRecurringRepo(recurringRepo)
	
	// Create test category
	category := test.CreateTestCategory(t, db, "Living", models.TransactionTypeExpense)
	
	// Create a recurring transaction
	recurring := &models.RecurringTransaction{
		Type:           models.TransactionTypeExpense,
		Amount:         1500.00,
		Currency:       "USD",
		CategoryID:     category.ID,
		Description:    "Rent",
		Frequency:      models.FrequencyMonthly,
		FrequencyValue: 1,
		StartDate:      time.Now().AddDate(0, -1, 0),
		NextDueDate:    time.Now(),
		IsActive:       true,
	}
	require.NoError(t, recurringRepo.Create(recurring))
	
	// Create transactions for current month
	now := time.Now()
	
	// One-time expense
	tx1 := &models.Transaction{
		Type:        models.TransactionTypeExpense,
		Amount:      250.00,
		Currency:    "USD",
		AmountUSD:   250.00,
		CategoryID:  category.ID,
		Description: "Groceries",
		Date:        now,
	}
	require.NoError(t, txRepo.Create(tx1))
	
	// Recurring expense
	tx2 := &models.Transaction{
		Type:                   models.TransactionTypeExpense,
		Amount:                 1500.00,
		Currency:               "USD",
		AmountUSD:              1500.00,
		CategoryID:             category.ID,
		Description:            "Rent",
		Date:                   now,
		RecurringTransactionID: &recurring.ID,
	}
	require.NoError(t, txRepo.Create(tx2))
	
	// Get burn rate
	burnRate, err := service.GetCurrentMonthBurnRate()
	require.NoError(t, err)
	require.NotNil(t, burnRate)
	
	// Verify calculations
	assert.Equal(t, 250.00, burnRate.OneTimeExpenses)
	assert.Equal(t, 1, burnRate.OneTimeCount)
	assert.Equal(t, 1500.00, burnRate.RecurringExpenses)
	assert.Equal(t, 1, burnRate.RecurringCount)
	assert.Equal(t, 1750.00, burnRate.TotalBurn)
	assert.Equal(t, 1500.00, burnRate.ProjectedMonthly)
	assert.Equal(t, 18000.00, burnRate.ProjectedYearly)
}