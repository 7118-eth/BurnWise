package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"burnwise/internal/models"
)

func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	testDataDir := "./test/data"
	if err := os.MkdirAll(testDataDir, 0755); err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}

	dbPath := filepath.Join(testDataDir, fmt.Sprintf("test_%s_%d.db", t.Name(), time.Now().UnixNano()))

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.Transaction{},
		&models.Category{},
		&models.Budget{},
		&models.CategoryHistory{},
		&models.RecurringTransaction{},
		&models.RecurringTransactionOccurrence{},
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
		os.Remove(dbPath)
	})

	return db
}

func SeedDefaultCategories(t *testing.T, db *gorm.DB) {
	t.Helper()

	categories := models.GetDefaultCategories()
	for _, category := range categories {
		err := db.Create(&category).Error
		require.NoError(t, err)
	}
}

func CleanupTestData(t *testing.T, db *gorm.DB) {
	t.Helper()

	db.Exec("DELETE FROM transactions")
	db.Exec("DELETE FROM categories WHERE is_default = false")
	db.Exec("DELETE FROM budgets")
}

func CreateTestCategory(t *testing.T, db *gorm.DB, name string, txType models.TransactionType) *models.Category {
	t.Helper()

	category := &models.Category{
		Name:  name,
		Type:  txType,
		Icon:  "💰",
		Color: "#4CAF50",
	}

	err := db.Create(category).Error
	require.NoError(t, err)

	return category
}

func CreateTestTransaction(t *testing.T, db *gorm.DB, amount float64, categoryID uint) *models.Transaction {
	t.Helper()

	tx := &models.Transaction{
		Type:        models.TransactionTypeExpense,
		Amount:      amount,
		Currency:    "USD",
		AmountUSD:   amount,
		CategoryID:  categoryID,
		Description: "Test transaction",
		Date:        time.Now(),
	}

	err := db.Create(tx).Error
	require.NoError(t, err)

	return tx
}

func CreateTestBudget(t *testing.T, db *gorm.DB, categoryID uint, amount float64) *models.Budget {
	t.Helper()

	budget := &models.Budget{
		Name:       "Test Budget",
		CategoryID: categoryID,
		Amount:     amount,
		Period:     models.BudgetPeriodMonthly,
		StartDate:  time.Now().AddDate(0, 0, -15),
	}

	err := db.Create(budget).Error
	require.NoError(t, err)

	return budget
}