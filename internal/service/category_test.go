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

func TestCategoryService_Create(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := repository.NewCategoryRepository(db)
	service := NewCategoryService(repo)
	
	category := &models.Category{
		Name:  "Groceries",
		Type:  models.TransactionTypeExpense,
		Icon:  "🛒",
		Color: "#FF5722",
	}
	
	err := service.Create(category)
	require.NoError(t, err)
	assert.Greater(t, category.ID, uint(0))
	
	// Try to create duplicate
	duplicate := &models.Category{
		Name: "Groceries",
		Type: models.TransactionTypeExpense,
		Icon: "🛒",
	}
	
	err = service.Create(duplicate)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestCategoryService_Delete(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := repository.NewCategoryRepository(db)
	service := NewCategoryService(repo)
	
	// Create category
	category := test.CreateTestCategory(t, db, "Test Category", models.TransactionTypeExpense)
	
	// Should delete successfully when no transactions
	err := service.Delete(category.ID)
	require.NoError(t, err)
	
	// Create another category with transaction
	category2 := test.CreateTestCategory(t, db, "Used Category", models.TransactionTypeExpense)
	test.CreateTestTransaction(t, db, 100.00, category2.ID)
	
	// Should not delete when has transactions
	err = service.Delete(category2.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete category with")
}

func TestCategoryService_GetWithTotals(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := repository.NewCategoryRepository(db)
	service := NewCategoryService(repo)
	
	// Create categories
	food := test.CreateTestCategory(t, db, "Food", models.TransactionTypeExpense)
	transport := test.CreateTestCategory(t, db, "Transport", models.TransactionTypeExpense)
	salary := test.CreateTestCategory(t, db, "Salary", models.TransactionTypeIncome)
	
	// Create transactions
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 1, 0).Add(-time.Second)
	
	test.CreateTestTransaction(t, db, 100.00, food.ID)
	test.CreateTestTransaction(t, db, 50.00, food.ID)
	test.CreateTestTransaction(t, db, 75.00, transport.ID)
	
	// Create income transaction
	tx := &models.Transaction{
		Type:        models.TransactionTypeIncome,
		Amount:      5000.00,
		Currency:    "USD",
		AmountUSD:   5000.00,
		CategoryID:  salary.ID,
		Description: "Monthly salary",
		Date:        time.Now(),
	}
	require.NoError(t, db.Create(tx).Error)
	
	// Get with totals
	totals, err := service.GetWithTotals(start, end)
	require.NoError(t, err)
	
	// Find categories in results
	var foodTotal, transportTotal, salaryTotal *models.CategoryWithTotal
	for _, cat := range totals {
		switch cat.ID {
		case food.ID:
			foodTotal = cat
		case transport.ID:
			transportTotal = cat
		case salary.ID:
			salaryTotal = cat
		}
	}
	
	require.NotNil(t, foodTotal)
	assert.Equal(t, 150.00, foodTotal.Total)
	assert.Equal(t, 2, foodTotal.Count)
	assert.InDelta(t, 66.67, foodTotal.Percentage, 0.01)
	
	require.NotNil(t, transportTotal)
	assert.Equal(t, 75.00, transportTotal.Total)
	assert.Equal(t, 1, transportTotal.Count)
	assert.InDelta(t, 33.33, transportTotal.Percentage, 0.01)
	
	require.NotNil(t, salaryTotal)
	assert.Equal(t, 5000.00, salaryTotal.Total)
	assert.Equal(t, 100.0, salaryTotal.Percentage)
}

func TestCategoryService_EnsureDefaultCategories(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := repository.NewCategoryRepository(db)
	service := NewCategoryService(repo)
	
	// Ensure defaults
	err := service.EnsureDefaultCategories()
	require.NoError(t, err)
	
	// Check they were created
	categories, err := service.GetAll()
	require.NoError(t, err)
	
	// Should have all default categories
	assert.GreaterOrEqual(t, len(categories), len(models.GetDefaultCategories()))
	
	// Run again - should not duplicate
	err = service.EnsureDefaultCategories()
	require.NoError(t, err)
	
	categories2, err := service.GetAll()
	require.NoError(t, err)
	assert.Equal(t, len(categories), len(categories2))
}