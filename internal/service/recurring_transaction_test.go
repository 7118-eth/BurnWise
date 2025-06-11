package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"budget-tracker/internal/models"
	"budget-tracker/internal/repository"
	test "budget-tracker/test/helpers"
)

func TestRecurringTransactionService_Create(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := repository.NewRecurringTransactionRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	
	tempDir := t.TempDir()
	settingsService, err := NewSettingsService(tempDir)
	require.NoError(t, err)
	currencyService := NewCurrencyService(settingsService)
	
	service := NewRecurringTransactionService(repo, txRepo, currencyService)
	
	// Create test category
	category := test.CreateTestCategory(t, db, "Rent", models.TransactionTypeExpense)

	// Create recurring transaction
	rt := &models.RecurringTransaction{
		Type:           models.TransactionTypeExpense,
		Amount:         1500.00,
		Currency:       "USD",
		CategoryID:     category.ID,
		Description:    "Monthly rent",
		Frequency:      models.FrequencyMonthly,
		FrequencyValue: 1,
		StartDate:      time.Now(),
		IsActive:       true,
	}

	err = service.Create(rt)
	require.NoError(t, err)
	assert.NotZero(t, rt.ID)
	assert.Equal(t, rt.StartDate, rt.NextDueDate)
}

func TestRecurringTransactionService_ProcessDueTransactions(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := repository.NewRecurringTransactionRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	
	tempDir := t.TempDir()
	settingsService, err := NewSettingsService(tempDir)
	require.NoError(t, err)
	currencyService := NewCurrencyService(settingsService)
	
	service := NewRecurringTransactionService(repo, txRepo, currencyService)
	
	// Create test category
	category := test.CreateTestCategory(t, db, "Utilities", models.TransactionTypeExpense)

	// Create recurring transaction due today
	today := time.Now()
	rt := &models.RecurringTransaction{
		Type:           models.TransactionTypeExpense,
		Amount:         100.00,
		Currency:       "USD",
		CategoryID:     category.ID,
		Description:    "Electricity bill",
		Frequency:      models.FrequencyMonthly,
		FrequencyValue: 1,
		StartDate:      today.AddDate(0, -1, 0), // Started a month ago
		NextDueDate:    today,                    // Due today
		IsActive:       true,
	}

	err = repo.Create(rt)
	require.NoError(t, err)

	// Process due transactions
	processed, err := service.ProcessDueTransactions(today)
	require.NoError(t, err)
	assert.Equal(t, 1, processed)

	// Verify transaction was created
	transactions, err := txRepo.GetAll()
	require.NoError(t, err)
	assert.Len(t, transactions, 1)

	tx := transactions[0]
	assert.Equal(t, rt.Amount, tx.Amount)
	assert.Equal(t, rt.Currency, tx.Currency)
	assert.Equal(t, rt.CategoryID, tx.CategoryID)
	assert.Equal(t, rt.Description, tx.Description)
	assert.NotNil(t, tx.RecurringTransactionID)
	assert.Equal(t, rt.ID, *tx.RecurringTransactionID)

	// Verify next due date was updated
	updatedRT, err := repo.GetByID(rt.ID)
	require.NoError(t, err)
	assert.True(t, updatedRT.NextDueDate.After(today))
}

func TestRecurringTransactionService_SkipOccurrence(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := repository.NewRecurringTransactionRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	
	tempDir := t.TempDir()
	settingsService, err := NewSettingsService(tempDir)
	require.NoError(t, err)
	currencyService := NewCurrencyService(settingsService)
	
	service := NewRecurringTransactionService(repo, txRepo, currencyService)
	
	// Create test category
	category := test.CreateTestCategory(t, db, "Subscription", models.TransactionTypeExpense)

	// Create recurring transaction
	today := time.Now()
	rt := &models.RecurringTransaction{
		Type:           models.TransactionTypeExpense,
		Amount:         9.99,
		Currency:       "USD",
		CategoryID:     category.ID,
		Description:    "Cloud hosting service",
		Frequency:      models.FrequencyMonthly,
		FrequencyValue: 1,
		StartDate:      today,
		NextDueDate:    today,
		IsActive:       true,
	}

	err = repo.Create(rt)
	require.NoError(t, err)

	// Skip today's occurrence
	err = service.SkipOccurrence(rt.ID, today, "Cancelled this month")
	require.NoError(t, err)

	// Process due transactions
	processed, err := service.ProcessDueTransactions(today)
	require.NoError(t, err)
	assert.Equal(t, 1, processed) // Processed but skipped

	// Verify no transaction was created (because it was skipped)
	transactions, err := txRepo.GetAll()
	require.NoError(t, err)
	assert.Len(t, transactions, 0)
}

func TestRecurringTransactionService_ModifyOccurrence(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := repository.NewRecurringTransactionRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	
	tempDir := t.TempDir()
	settingsService, err := NewSettingsService(tempDir)
	require.NoError(t, err)
	currencyService := NewCurrencyService(settingsService)
	
	service := NewRecurringTransactionService(repo, txRepo, currencyService)
	
	// Create test category
	category := test.CreateTestCategory(t, db, "Salary", models.TransactionTypeIncome)

	// Create recurring transaction
	today := time.Now()
	rt := &models.RecurringTransaction{
		Type:           models.TransactionTypeIncome,
		Amount:         5000.00,
		Currency:       "USD",
		CategoryID:     category.ID,
		Description:    "Monthly salary",
		Frequency:      models.FrequencyMonthly,
		FrequencyValue: 1,
		StartDate:      today,
		NextDueDate:    today,
		IsActive:       true,
	}

	err = repo.Create(rt)
	require.NoError(t, err)

	// Modify today's occurrence
	modifiedAmount := 5500.00
	modifiedDesc := "Monthly salary + bonus"
	err = service.ModifyOccurrence(rt.ID, today, &modifiedAmount, &modifiedDesc)
	require.NoError(t, err)

	// Process due transactions
	processed, err := service.ProcessDueTransactions(today)
	require.NoError(t, err)
	assert.Equal(t, 1, processed)

	// Verify modified transaction was created
	transactions, err := txRepo.GetAll()
	require.NoError(t, err)
	require.Len(t, transactions, 1)

	tx := transactions[0]
	assert.Equal(t, modifiedAmount, tx.Amount)
	assert.Equal(t, modifiedDesc, tx.Description)
}

func TestRecurringTransactionService_PauseResume(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := repository.NewRecurringTransactionRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	
	tempDir := t.TempDir()
	settingsService, err := NewSettingsService(tempDir)
	require.NoError(t, err)
	currencyService := NewCurrencyService(settingsService)
	
	service := NewRecurringTransactionService(repo, txRepo, currencyService)
	
	// Create test category
	category := test.CreateTestCategory(t, db, "Insurance", models.TransactionTypeExpense)

	// Create recurring transaction
	rt := &models.RecurringTransaction{
		Type:           models.TransactionTypeExpense,
		Amount:         200.00,
		Currency:       "USD",
		CategoryID:     category.ID,
		Description:    "Car insurance",
		Frequency:      models.FrequencyMonthly,
		FrequencyValue: 1,
		StartDate:      time.Now(),
		NextDueDate:    time.Now(), // Set NextDueDate
		IsActive:       true,
	}

	// Verify the model has the correct amount before creating
	assert.Equal(t, 200.00, rt.Amount)
	
	err = service.Create(rt)
	require.NoError(t, err, "Failed to create recurring transaction")
	assert.True(t, rt.IsActive)

	// Pause
	err = service.Pause(rt.ID)
	require.NoError(t, err)

	// Verify paused
	paused, err := repo.GetByID(rt.ID)
	require.NoError(t, err)
	assert.False(t, paused.IsActive)

	// Resume
	err = service.Resume(rt.ID)
	require.NoError(t, err)

	// Verify resumed
	resumed, err := repo.GetByID(rt.ID)
	require.NoError(t, err)
	assert.True(t, resumed.IsActive)
}

func TestRecurringTransactionService_EndDateHandling(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := repository.NewRecurringTransactionRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	
	tempDir := t.TempDir()
	settingsService, err := NewSettingsService(tempDir)
	require.NoError(t, err)
	currencyService := NewCurrencyService(settingsService)
	
	service := NewRecurringTransactionService(repo, txRepo, currencyService)
	
	// Create test category
	category := test.CreateTestCategory(t, db, "Loan", models.TransactionTypeExpense)

	// Create recurring transaction with end date
	today := time.Now()
	yesterday := today.AddDate(0, 0, -1)
	rt := &models.RecurringTransaction{
		Type:           models.TransactionTypeExpense,
		Amount:         500.00,
		Currency:       "USD",
		CategoryID:     category.ID,
		Description:    "Loan payment",
		Frequency:      models.FrequencyMonthly,
		FrequencyValue: 1,
		StartDate:      today.AddDate(0, -2, 0),
		EndDate:        &yesterday, // Ended yesterday
		NextDueDate:    today,
		IsActive:       true,
	}

	err = repo.Create(rt)
	require.NoError(t, err)

	// Process due transactions
	processed, err := service.ProcessDueTransactions(today)
	require.NoError(t, err)
	assert.Equal(t, 0, processed) // Should not process as it's past end date

	// Verify no transaction was created
	transactions, err := txRepo.GetAll()
	require.NoError(t, err)
	assert.Len(t, transactions, 0)
}

func TestRecurringTransactionService_GetUpcoming(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := repository.NewRecurringTransactionRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	
	tempDir := t.TempDir()
	settingsService, err := NewSettingsService(tempDir)
	require.NoError(t, err)
	currencyService := NewCurrencyService(settingsService)
	
	service := NewRecurringTransactionService(repo, txRepo, currencyService)
	
	// Create test category
	category := test.CreateTestCategory(t, db, "Bills", models.TransactionTypeExpense)

	// Create recurring transactions
	today := time.Now()
	
	// Due in 5 days
	rt1 := &models.RecurringTransaction{
		Type:           models.TransactionTypeExpense,
		Amount:         50.00,
		Currency:       "USD",
		CategoryID:     category.ID,
		Description:    "Internet bill",
		Frequency:      models.FrequencyMonthly,
		FrequencyValue: 1,
		StartDate:      today,
		NextDueDate:    today.AddDate(0, 0, 5),
		IsActive:       true,
	}
	err = repo.Create(rt1)
	require.NoError(t, err)

	// Due in 10 days
	rt2 := &models.RecurringTransaction{
		Type:           models.TransactionTypeExpense,
		Amount:         100.00,
		Currency:       "USD",
		CategoryID:     category.ID,
		Description:    "Phone bill",
		Frequency:      models.FrequencyMonthly,
		FrequencyValue: 1,
		StartDate:      today,
		NextDueDate:    today.AddDate(0, 0, 10),
		IsActive:       true,
	}
	err = repo.Create(rt2)
	require.NoError(t, err)

	// Due in 20 days (outside range)
	rt3 := &models.RecurringTransaction{
		Type:           models.TransactionTypeExpense,
		Amount:         150.00,
		Currency:       "USD",
		CategoryID:     category.ID,
		Description:    "Electricity bill",
		Frequency:      models.FrequencyMonthly,
		FrequencyValue: 1,
		StartDate:      today,
		NextDueDate:    today.AddDate(0, 0, 20),
		IsActive:       true,
	}
	err = repo.Create(rt3)
	require.NoError(t, err)

	// Get upcoming in next 14 days
	upcoming, err := service.GetUpcoming(14)
	require.NoError(t, err)
	assert.Len(t, upcoming, 2)
}

func TestRecurringTransactionService_CalculateProjectedAmount(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := repository.NewRecurringTransactionRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	
	tempDir := t.TempDir()
	settingsService, err := NewSettingsService(tempDir)
	require.NoError(t, err)
	currencyService := NewCurrencyService(settingsService)
	
	service := NewRecurringTransactionService(repo, txRepo, currencyService)
	
	// Create test categories
	incomeCategory := test.CreateTestCategory(t, db, "Salary", models.TransactionTypeIncome)
	expenseCategory := test.CreateTestCategory(t, db, "Rent", models.TransactionTypeExpense)

	// Create monthly income
	income := &models.RecurringTransaction{
		Type:           models.TransactionTypeIncome,
		Amount:         5000.00,
		Currency:       "USD",
		CategoryID:     incomeCategory.ID,
		Description:    "Monthly salary",
		Frequency:      models.FrequencyMonthly,
		FrequencyValue: 1,
		StartDate:      time.Now().AddDate(0, -6, 0),
		NextDueDate:    time.Now(),
		IsActive:       true,
	}
	err = repo.Create(income)
	require.NoError(t, err)

	// Create monthly expense
	expense := &models.RecurringTransaction{
		Type:           models.TransactionTypeExpense,
		Amount:         1500.00,
		Currency:       "USD",
		CategoryID:     expenseCategory.ID,
		Description:    "Monthly rent",
		Frequency:      models.FrequencyMonthly,
		FrequencyValue: 1,
		StartDate:      time.Now().AddDate(0, -6, 0),
		NextDueDate:    time.Now(),
		IsActive:       true,
	}
	err = repo.Create(expense)
	require.NoError(t, err)

	// Calculate projection for next 3 months
	startDate := time.Now()
	endDate := startDate.AddDate(0, 3, 0)

	projected, err := service.CalculateProjectedAmount(startDate, endDate)
	require.NoError(t, err)

	// Should be (5000 - 1500) * 3 = 10500
	assert.Equal(t, 10500.0, projected)
}