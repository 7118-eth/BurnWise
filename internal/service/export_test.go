package service

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"budget-tracker/internal/models"
	"budget-tracker/internal/repository"
	test "budget-tracker/test/helpers"
)

func TestExportService_ExportTransactionsCSV(t *testing.T) {
	db := test.SetupTestDB(t)
	txRepo := repository.NewTransactionRepository(db)
	
	tempDir := t.TempDir()
	settingsService, err := NewSettingsService(tempDir)
	require.NoError(t, err)
	currencyService := NewCurrencyService(settingsService)
	
	txService := NewTransactionService(txRepo, currencyService)
	exportService := NewExportService(txService)
	
	// Create test data
	category := test.CreateTestCategory(t, db, "Food", models.TransactionTypeExpense)
	
	tx1 := &models.Transaction{
		Type:        models.TransactionTypeExpense,
		Amount:      50.00,
		Currency:    "USD",
		CategoryID:  category.ID,
		Description: "Groceries",
		Date:        time.Now(),
	}
	require.NoError(t, txService.Create(tx1))
	
	tx2 := &models.Transaction{
		Type:        models.TransactionTypeExpense,
		Amount:      100.00,
		Currency:    "AED",
		CategoryID:  category.ID,
		Description: "Restaurant",
		Date:        time.Now(),
	}
	require.NoError(t, txService.Create(tx2))
	
	// Export to buffer
	var buf bytes.Buffer
	err = exportService.ExportTransactionsCSV(&buf, &models.TransactionFilter{})
	require.NoError(t, err)
	
	// Parse CSV
	reader := csv.NewReader(strings.NewReader(buf.String()))
	records, err := reader.ReadAll()
	require.NoError(t, err)
	
	// Check header
	assert.Len(t, records, 3) // header + 2 transactions
	assert.Equal(t, []string{"Date", "Type", "Category", "Description", "Amount", "Currency", "Amount (USD)"}, records[0])
	
	// Check both transactions are present (order may vary)
	var groceriesFound, restaurantFound bool
	for i := 1; i < len(records); i++ {
		assert.Equal(t, "expense", records[i][1])
		assert.Equal(t, "Food", records[i][2])
		
		if records[i][3] == "Groceries" {
			groceriesFound = true
			assert.Equal(t, "50.00", records[i][4])
			assert.Equal(t, "USD", records[i][5])
			assert.Equal(t, "50.00", records[i][6])
		} else if records[i][3] == "Restaurant" {
			restaurantFound = true
			assert.Equal(t, "100.00", records[i][4])
			assert.Equal(t, "AED", records[i][5])
			assert.Contains(t, records[i][6], "27.2") // Converted amount
		}
	}
	assert.True(t, groceriesFound, "Groceries transaction not found")
	assert.True(t, restaurantFound, "Restaurant transaction not found")
}

func TestExportService_ExportMonthlyReportCSV(t *testing.T) {
	db := test.SetupTestDB(t)
	txRepo := repository.NewTransactionRepository(db)
	
	tempDir := t.TempDir()
	settingsService, err := NewSettingsService(tempDir)
	require.NoError(t, err)
	currencyService := NewCurrencyService(settingsService)
	
	txService := NewTransactionService(txRepo, currencyService)
	exportService := NewExportService(txService)
	
	// Create test data
	incomeCategory := test.CreateTestCategory(t, db, "Salary", models.TransactionTypeIncome)
	expenseCategory := test.CreateTestCategory(t, db, "Food", models.TransactionTypeExpense)
	
	// Create income
	income := &models.Transaction{
		Type:        models.TransactionTypeIncome,
		Amount:      5000.00,
		Currency:    "USD",
		CategoryID:  incomeCategory.ID,
		Description: "Monthly salary",
		Date:        time.Now(),
	}
	require.NoError(t, txService.Create(income))
	
	// Create expenses
	expense := &models.Transaction{
		Type:        models.TransactionTypeExpense,
		Amount:      100.00,
		Currency:    "USD",
		CategoryID:  expenseCategory.ID,
		Description: "Groceries",
		Date:        time.Now(),
	}
	require.NoError(t, txService.Create(expense))
	
	// Export report
	var buf bytes.Buffer
	err = exportService.ExportMonthlyReportCSV(&buf, time.Now().Year(), time.Now().Month())
	require.NoError(t, err)
	
	// Check output contains expected data
	output := buf.String()
	assert.Contains(t, output, "Monthly Report")
	assert.Contains(t, output, "Total Income,5000.00")
	assert.Contains(t, output, "Total Expenses,100.00")
	assert.Contains(t, output, "Balance,4900.00")
	assert.Contains(t, output, "Category Breakdown")
	assert.Contains(t, output, "Salary")
	assert.Contains(t, output, "Food")
}

func TestExportService_ExportBudgetStatusCSV(t *testing.T) {
	db := test.SetupTestDB(t)
	budgetRepo := repository.NewBudgetRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	
	tempDir := t.TempDir()
	settingsService, err := NewSettingsService(tempDir)
	require.NoError(t, err)
	currencyService := NewCurrencyService(settingsService)
	
	txService := NewTransactionService(txRepo, currencyService)
	budgetService := NewBudgetService(budgetRepo, txRepo)
	exportService := NewExportService(txService)
	
	// Create test data
	category := test.CreateTestCategory(t, db, "Food", models.TransactionTypeExpense)
	test.CreateTestBudget(t, db, category.ID, 500.00)
	
	// Create transaction
	tx := &models.Transaction{
		Type:        models.TransactionTypeExpense,
		Amount:      100.00,
		Currency:    "USD",
		CategoryID:  category.ID,
		Description: "Groceries",
		Date:        time.Now(),
	}
	require.NoError(t, txService.Create(tx))
	
	// Export budget status
	var buf bytes.Buffer
	err = exportService.ExportBudgetStatusCSV(&buf, budgetService)
	require.NoError(t, err)
	
	// Parse CSV
	reader := csv.NewReader(strings.NewReader(buf.String()))
	records, err := reader.ReadAll()
	require.NoError(t, err)
	
	// Check data
	assert.Len(t, records, 2) // header + 1 budget
	assert.Equal(t, "Test Budget", records[1][0])
	assert.Equal(t, "Food", records[1][1])
	assert.Equal(t, "monthly", records[1][2])
	assert.Equal(t, "500.00", records[1][3])
	assert.Equal(t, "100.00", records[1][4])
	assert.Equal(t, "400.00", records[1][5])
	assert.Contains(t, records[1][6], "20.0%")
	assert.Equal(t, "OK", records[1][7])
}