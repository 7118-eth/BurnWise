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

func TestTransactionService_Create(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := repository.NewTransactionRepository(db)
	currencyService := NewCurrencyService()
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
	
	err := service.Create(tx)
	require.NoError(t, err)
	assert.Greater(t, tx.ID, uint(0))
	assert.Equal(t, tx.Amount, tx.AmountUSD)
}

func TestTransactionService_CreateWithCurrency(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := repository.NewTransactionRepository(db)
	currencyService := NewCurrencyService()
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
	
	err := service.Create(tx)
	require.NoError(t, err)
	assert.Greater(t, tx.ID, uint(0))
	assert.InDelta(t, 27.23, tx.AmountUSD, 0.01)
}

func TestTransactionService_GetCurrentMonthSummary(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := repository.NewTransactionRepository(db)
	currencyService := NewCurrencyService()
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