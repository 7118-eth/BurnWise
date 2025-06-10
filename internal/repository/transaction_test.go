package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"budget-tracker/internal/models"
	test "budget-tracker/test/helpers"
	"budget-tracker/test/fixtures"
)

func TestTransactionRepository_Create(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := NewTransactionRepository(db)
	
	category := test.CreateTestCategory(t, db, "Food", models.TransactionTypeExpense)
	
	tx := fixtures.NewTransaction().
		WithCategory(category.ID).
		WithAmount(50.00).
		Build()
	
	err := repo.Create(tx)
	require.NoError(t, err)
	assert.Greater(t, tx.ID, uint(0))
	
	found, err := repo.GetByID(tx.ID)
	require.NoError(t, err)
	assert.Equal(t, tx.Description, found.Description)
	assert.Equal(t, tx.Amount, found.Amount)
}

func TestTransactionRepository_GetByDateRange(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := NewTransactionRepository(db)
	
	category := test.CreateTestCategory(t, db, "Food", models.TransactionTypeExpense)
	
	yesterday := fixtures.NewTransaction().
		WithCategory(category.ID).
		WithDate(time.Now().AddDate(0, 0, -1)).
		WithDescription("Yesterday").
		Build()
	today := fixtures.NewTransaction().
		WithCategory(category.ID).
		WithDate(time.Now()).
		WithDescription("Today").
		Build()
	tomorrow := fixtures.NewTransaction().
		WithCategory(category.ID).
		WithDate(time.Now().AddDate(0, 0, 1)).
		WithDescription("Tomorrow").
		Build()
	
	require.NoError(t, repo.Create(yesterday))
	require.NoError(t, repo.Create(today))
	require.NoError(t, repo.Create(tomorrow))
	
	// Set specific times to ensure proper date boundaries
	startOfYesterday := time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour)
	endOfToday := time.Now().Truncate(24 * time.Hour).Add(24*time.Hour - time.Second)
	
	results, err := repo.GetByDateRange(startOfYesterday, endOfToday)
	
	require.NoError(t, err)
	assert.Len(t, results, 2)
	if len(results) >= 2 {
		assert.Equal(t, "Today", results[0].Description)
		assert.Equal(t, "Yesterday", results[1].Description)
	}
}

func TestTransactionRepository_GetByFilter(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := NewTransactionRepository(db)
	
	incomeCategory := test.CreateTestCategory(t, db, "Salary", models.TransactionTypeIncome)
	expenseCategory := test.CreateTestCategory(t, db, "Food", models.TransactionTypeExpense)
	
	income := fixtures.NewTransaction().
		WithType(models.TransactionTypeIncome).
		WithCategory(incomeCategory.ID).
		WithAmount(5000).
		WithDescription("Monthly salary").
		Build()
	
	expense1 := fixtures.NewTransaction().
		WithType(models.TransactionTypeExpense).
		WithCategory(expenseCategory.ID).
		WithAmount(50).
		WithDescription("Lunch at cafe").
		Build()
	
	expense2 := fixtures.NewTransaction().
		WithType(models.TransactionTypeExpense).
		WithCategory(expenseCategory.ID).
		WithAmount(150).
		WithDescription("Dinner at restaurant").
		Build()
	
	require.NoError(t, repo.Create(income))
	require.NoError(t, repo.Create(expense1))
	require.NoError(t, repo.Create(expense2))
	
	tests := []struct {
		name      string
		filter    *models.TransactionFilter
		wantCount int
	}{
		{
			name: "filter by type",
			filter: &models.TransactionFilter{
				Type: models.TransactionTypeExpense,
			},
			wantCount: 2,
		},
		{
			name: "filter by category",
			filter: &models.TransactionFilter{
				CategoryID: expenseCategory.ID,
			},
			wantCount: 2,
		},
		{
			name: "filter by amount range",
			filter: &models.TransactionFilter{
				MinAmount: 100,
				MaxAmount: 200,
			},
			wantCount: 1,
		},
		{
			name: "filter by search",
			filter: &models.TransactionFilter{
				Search: "restaurant",
			},
			wantCount: 1,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := repo.GetByFilter(tt.filter)
			require.NoError(t, err)
			assert.Len(t, results, tt.wantCount)
		})
	}
}

func TestTransactionRepository_GetSummary(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := NewTransactionRepository(db)
	
	incomeCategory := test.CreateTestCategory(t, db, "Salary", models.TransactionTypeIncome)
	expenseCategory := test.CreateTestCategory(t, db, "Food", models.TransactionTypeExpense)
	
	income := fixtures.NewTransaction().
		WithType(models.TransactionTypeIncome).
		WithCategory(incomeCategory.ID).
		WithAmount(5000).
		Build()
	
	expense1 := fixtures.NewTransaction().
		WithType(models.TransactionTypeExpense).
		WithCategory(expenseCategory.ID).
		WithAmount(100).
		Build()
	
	expense2 := fixtures.NewTransaction().
		WithType(models.TransactionTypeExpense).
		WithCategory(expenseCategory.ID).
		WithAmount(200).
		Build()
	
	require.NoError(t, repo.Create(income))
	require.NoError(t, repo.Create(expense1))
	require.NoError(t, repo.Create(expense2))
	
	start := time.Now().AddDate(0, 0, -7)
	end := time.Now().AddDate(0, 0, 1)
	
	summary, err := repo.GetSummary(start, end)
	require.NoError(t, err)
	
	assert.Equal(t, 5000.0, summary.TotalIncome)
	assert.Equal(t, 300.0, summary.TotalExpenses)
	assert.Equal(t, 4700.0, summary.Balance)
	assert.Equal(t, 3, summary.Count)
}

func TestTransactionRepository_GetCategorySummary(t *testing.T) {
	db := test.SetupTestDB(t)
	repo := NewTransactionRepository(db)
	
	foodCategory := test.CreateTestCategory(t, db, "Food", models.TransactionTypeExpense)
	transportCategory := test.CreateTestCategory(t, db, "Transport", models.TransactionTypeExpense)
	
	require.NoError(t, repo.Create(fixtures.NewTransaction().
		WithCategory(foodCategory.ID).
		WithAmount(100).
		Build()))
	
	require.NoError(t, repo.Create(fixtures.NewTransaction().
		WithCategory(foodCategory.ID).
		WithAmount(50).
		Build()))
	
	require.NoError(t, repo.Create(fixtures.NewTransaction().
		WithCategory(transportCategory.ID).
		WithAmount(75).
		Build()))
	
	start := time.Now().AddDate(0, 0, -7)
	end := time.Now().AddDate(0, 0, 1)
	
	summary, err := repo.GetCategorySummary(start, end)
	require.NoError(t, err)
	
	assert.Len(t, summary, 2)
	assert.Equal(t, "Food", summary[0].Name)
	assert.Equal(t, 150.0, summary[0].Total)
	assert.Equal(t, 2, summary[0].Count)
	assert.InDelta(t, 66.67, summary[0].Percentage, 0.01)
	
	assert.Equal(t, "Transport", summary[1].Name)
	assert.Equal(t, 75.0, summary[1].Total)
	assert.Equal(t, 1, summary[1].Count)
	assert.InDelta(t, 33.33, summary[1].Percentage, 0.01)
}