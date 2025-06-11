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

func TestCategoryService_Update_WithHistory(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := repository.NewCategoryRepository(db)
	service := NewCategoryService(repo)

	// Create a test category
	category := &models.Category{
		Name:  "Original Food",
		Type:  models.TransactionTypeExpense,
		Icon:  "🍕",
		Color: "#FF5722",
	}
	err := service.Create(category)
	require.NoError(t, err)

	// Update the category
	category.Name = "Updated Food"
	category.Icon = "🍔"
	category.Color = "#4CAF50"

	err = service.Update(category)
	require.NoError(t, err)

	// Check that history was recorded
	history, err := service.GetHistory(category.ID)
	require.NoError(t, err)
	assert.Len(t, history, 1)

	historyRecord := history[0]
	assert.Equal(t, models.CategoryActionEdited, historyRecord.Action)
	assert.Equal(t, "Original Food", historyRecord.OldName)
	assert.Equal(t, "Updated Food", historyRecord.NewName)
	assert.Equal(t, "🍕", historyRecord.OldIcon)
	assert.Equal(t, "🍔", historyRecord.NewIcon)
	assert.Equal(t, "#FF5722", historyRecord.OldColor)
	assert.Equal(t, "#4CAF50", historyRecord.NewColor)
}

func TestCategoryService_MergeCategories(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := repository.NewCategoryRepository(db)
	service := NewCategoryService(repo)

	tempDir := t.TempDir()
	settingsService, err := NewSettingsService(tempDir)
	require.NoError(t, err)
	currencyService := NewCurrencyService(settingsService)
	txRepo := repository.NewTransactionRepository(db)
	txService := NewTransactionService(txRepo, currencyService)

	// Create source and target categories
	sourceCategory := &models.Category{
		Name:  "Fast Food",
		Type:  models.TransactionTypeExpense,
		Icon:  "🍔",
		Color: "#FF5722",
	}
	err = service.Create(sourceCategory)
	require.NoError(t, err)

	targetCategory := &models.Category{
		Name:  "Food & Dining",
		Type:  models.TransactionTypeExpense,
		Icon:  "🍽️",
		Color: "#4CAF50",
	}
	err = service.Create(targetCategory)
	require.NoError(t, err)

	// Create transactions in source category
	tx1 := &models.Transaction{
		Type:        models.TransactionTypeExpense,
		Amount:      25.00,
		Currency:    "USD",
		CategoryID:  sourceCategory.ID,
		Description: "McDonald's",
		Date:        time.Now(),
	}
	err = txService.Create(tx1)
	require.NoError(t, err)

	tx2 := &models.Transaction{
		Type:        models.TransactionTypeExpense,
		Amount:      15.00,
		Currency:    "USD",
		CategoryID:  sourceCategory.ID,
		Description: "Burger King",
		Date:        time.Now(),
	}
	err = txService.Create(tx2)
	require.NoError(t, err)

	// Perform merge
	err = service.MergeCategories(sourceCategory.ID, targetCategory.ID)
	require.NoError(t, err)

	// Verify source category is deleted
	_, err = service.GetByID(sourceCategory.ID)
	assert.Error(t, err)

	// Verify transactions are moved to target category
	tx1Updated, err := txRepo.GetByID(tx1.ID)
	require.NoError(t, err)
	assert.Equal(t, targetCategory.ID, tx1Updated.CategoryID)

	tx2Updated, err := txRepo.GetByID(tx2.ID)
	require.NoError(t, err)
	assert.Equal(t, targetCategory.ID, tx2Updated.CategoryID)

	// Verify history was recorded
	history, err := service.GetHistory(sourceCategory.ID)
	require.NoError(t, err)
	assert.Len(t, history, 1)

	historyRecord := history[0]
	assert.Equal(t, models.CategoryActionMerged, historyRecord.Action)
	assert.Equal(t, "Fast Food", historyRecord.OldName)
	assert.Equal(t, targetCategory.ID, *historyRecord.TargetCategoryID)
	assert.Equal(t, 2, historyRecord.TransactionCount)
}

func TestCategoryService_MergeCategories_DifferentTypes(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := repository.NewCategoryRepository(db)
	service := NewCategoryService(repo)

	// Create categories of different types
	incomeCategory := &models.Category{
		Name: "Salary",
		Type: models.TransactionTypeIncome,
		Icon: "💼",
	}
	err := service.Create(incomeCategory)
	require.NoError(t, err)

	expenseCategory := &models.Category{
		Name: "Food",
		Type: models.TransactionTypeExpense,
		Icon: "🍔",
	}
	err = service.Create(expenseCategory)
	require.NoError(t, err)

	// Attempt to merge different types
	err = service.MergeCategories(incomeCategory.ID, expenseCategory.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot merge categories of different types")
}

func TestCategoryService_MergeCategories_DefaultCategory(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := repository.NewCategoryRepository(db)
	service := NewCategoryService(repo)

	// Create a default category and a custom category
	defaultCategory := &models.Category{
		Name:      "Salary",
		Type:      models.TransactionTypeIncome,
		Icon:      "💼",
		IsDefault: true,
	}
	err := repo.Create(defaultCategory)
	require.NoError(t, err)

	customCategory := &models.Category{
		Name: "Freelance",
		Type: models.TransactionTypeIncome,
		Icon: "💻",
	}
	err = service.Create(customCategory)
	require.NoError(t, err)

	// Attempt to merge default category
	err = service.MergeCategories(defaultCategory.ID, customCategory.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot merge default category")
}

func TestCategoryService_GetAllWithUsageCount(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := repository.NewCategoryRepository(db)
	service := NewCategoryService(repo)

	tempDir := t.TempDir()
	settingsService, err := NewSettingsService(tempDir)
	require.NoError(t, err)
	currencyService := NewCurrencyService(settingsService)
	txRepo := repository.NewTransactionRepository(db)
	txService := NewTransactionService(txRepo, currencyService)

	// Create test category
	category := &models.Category{
		Name: "Test Food",
		Type: models.TransactionTypeExpense,
		Icon: "🍔",
	}
	err = service.Create(category)
	require.NoError(t, err)

	// Create transactions
	for i := 0; i < 3; i++ {
		tx := &models.Transaction{
			Type:        models.TransactionTypeExpense,
			Amount:      10.00,
			Currency:    "USD",
			CategoryID:  category.ID,
			Description: "Test transaction",
			Date:        time.Now(),
		}
		err = txService.Create(tx)
		require.NoError(t, err)
	}

	// Get categories with usage count
	categories, err := service.GetAllWithUsageCount()
	require.NoError(t, err)

	// Find our test category
	var testCategory *models.CategoryWithTotal
	for _, cat := range categories {
		if cat.ID == category.ID {
			testCategory = cat
			break
		}
	}

	require.NotNil(t, testCategory)
	assert.Equal(t, 3, testCategory.Count)
}

func TestCategoryService_Delete_PreventWithTransactions(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := repository.NewCategoryRepository(db)
	service := NewCategoryService(repo)

	tempDir := t.TempDir()
	settingsService, err := NewSettingsService(tempDir)
	require.NoError(t, err)
	currencyService := NewCurrencyService(settingsService)
	txRepo := repository.NewTransactionRepository(db)
	txService := NewTransactionService(txRepo, currencyService)

	// Create test category
	category := &models.Category{
		Name: "Test Category",
		Type: models.TransactionTypeExpense,
		Icon: "📁",
	}
	err = service.Create(category)
	require.NoError(t, err)

	// Create a transaction
	tx := &models.Transaction{
		Type:        models.TransactionTypeExpense,
		Amount:      10.00,
		Currency:    "USD",
		CategoryID:  category.ID,
		Description: "Test transaction",
		Date:        time.Now(),
	}
	err = txService.Create(tx)
	require.NoError(t, err)

	// Attempt to delete category with transactions
	err = service.Delete(category.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete category with")

	// Verify category still exists
	_, err = service.GetByID(category.ID)
	assert.NoError(t, err)
}

func TestCategoryService_Delete_PreventDefault(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := repository.NewCategoryRepository(db)
	service := NewCategoryService(repo)

	// Create a default category
	defaultCategory := &models.Category{
		Name:      "Default Food",
		Type:      models.TransactionTypeExpense,
		Icon:      "🍔",
		IsDefault: true,
	}
	err := repo.Create(defaultCategory)
	require.NoError(t, err)

	// Attempt to delete default category
	err = service.Delete(defaultCategory.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete default category")

	// Verify category still exists
	_, err = service.GetByID(defaultCategory.ID)
	assert.NoError(t, err)
}