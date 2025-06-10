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

func TestBudgetService_Create(t *testing.T) {
	db := test.SetupTestDB(t)
	budgetRepo := repository.NewBudgetRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	service := NewBudgetService(budgetRepo, txRepo)
	
	category := test.CreateTestCategory(t, db, "Food", models.TransactionTypeExpense)
	
	budget := &models.Budget{
		Name:       "Food Budget",
		CategoryID: category.ID,
		Amount:     500.00,
		Period:     models.BudgetPeriodMonthly,
		StartDate:  time.Now(),
	}
	
	err := service.Create(budget)
	require.NoError(t, err)
	assert.Greater(t, budget.ID, uint(0))
	
	// Try to create duplicate active budget
	duplicate := &models.Budget{
		Name:       "Another Food Budget",
		CategoryID: category.ID,
		Amount:     600.00,
		Period:     models.BudgetPeriodMonthly,
		StartDate:  time.Now(),
	}
	
	err = service.Create(duplicate)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "active budget already exists")
}

func TestBudgetService_GetStatus(t *testing.T) {
	db := test.SetupTestDB(t)
	budgetRepo := repository.NewBudgetRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	service := NewBudgetService(budgetRepo, txRepo)
	
	category := test.CreateTestCategory(t, db, "Food", models.TransactionTypeExpense)
	
	// Create budget
	budget := test.CreateTestBudget(t, db, category.ID, 1000.00)
	
	// Create some transactions
	for i := 0; i < 3; i++ {
		tx := &models.Transaction{
			Type:        models.TransactionTypeExpense,
			Amount:      200.00,
			Currency:    "USD",
			AmountUSD:   200.00,
			CategoryID:  category.ID,
			Description: "Test expense",
			Date:        time.Now(),
		}
		require.NoError(t, db.Create(tx).Error)
	}
	
	// Get status
	status, err := service.GetStatus(budget.ID)
	require.NoError(t, err)
	
	assert.Equal(t, 600.00, status.Spent)
	assert.Equal(t, 400.00, status.Remaining)
	assert.Equal(t, 60.0, status.PercentUsed)
	assert.False(t, status.IsOverBudget)
}

func TestBudgetService_CheckOverspending(t *testing.T) {
	db := test.SetupTestDB(t)
	budgetRepo := repository.NewBudgetRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	service := NewBudgetService(budgetRepo, txRepo)
	
	category := test.CreateTestCategory(t, db, "Shopping", models.TransactionTypeExpense)
	
	// Create budget with low amount
	budget := test.CreateTestBudget(t, db, category.ID, 100.00)
	
	// Create transaction that exceeds budget
	tx := &models.Transaction{
		Type:        models.TransactionTypeExpense,
		Amount:      150.00,
		Currency:    "USD",
		AmountUSD:   150.00,
		CategoryID:  category.ID,
		Description: "Big purchase",
		Date:        time.Now(),
	}
	require.NoError(t, db.Create(tx).Error)
	
	// Check overspending
	isOver, amount, err := service.CheckOverspending(budget.ID)
	require.NoError(t, err)
	
	assert.True(t, isOver)
	assert.Equal(t, 50.00, amount)
}

func TestBudgetService_GetAllStatuses(t *testing.T) {
	db := test.SetupTestDB(t)
	budgetRepo := repository.NewBudgetRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	service := NewBudgetService(budgetRepo, txRepo)
	
	// Create multiple budgets
	cat1 := test.CreateTestCategory(t, db, "Food", models.TransactionTypeExpense)
	cat2 := test.CreateTestCategory(t, db, "Transport", models.TransactionTypeExpense)
	
	test.CreateTestBudget(t, db, cat1.ID, 500.00)
	test.CreateTestBudget(t, db, cat2.ID, 300.00)
	
	// Create some transactions
	test.CreateTestTransaction(t, db, 100.00, cat1.ID)
	test.CreateTestTransaction(t, db, 50.00, cat2.ID)
	
	// Get all statuses
	statuses, err := service.GetAllStatuses()
	require.NoError(t, err)
	
	assert.Len(t, statuses, 2)
	
	// Check first budget status
	assert.Equal(t, 100.00, statuses[0].Spent)
	assert.Equal(t, 20.0, statuses[0].PercentUsed)
	
	// Check second budget status
	assert.Equal(t, 50.00, statuses[1].Spent)
	assert.InDelta(t, 16.67, statuses[1].PercentUsed, 0.01)
}