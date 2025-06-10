package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"budget-tracker/internal/models"
	"budget-tracker/internal/repository"
	test "budget-tracker/test/helpers"
)

func TestSettingsService(t *testing.T) {
	// Create temp directory for settings file
	tempDir := t.TempDir()

	t.Run("NewSettingsService creates default settings", func(t *testing.T) {
		service, err := NewSettingsService(tempDir)
		require.NoError(t, err)
		assert.NotNil(t, service)

		// Check default settings
		settings := service.Get()
		assert.Equal(t, []string{"USD", "EUR", "AED"}, settings.Currencies.Enabled)
		assert.Equal(t, "USD", settings.Currencies.Default)
		assert.Equal(t, 3.6725, settings.Currencies.FixedRates["AED"])
		assert.Equal(t, "2006-01-02", settings.UI.DateFormat)
		assert.Equal(t, 2, settings.UI.DecimalPlaces)

		// Check that file was created
		settingsPath := filepath.Join(tempDir, "settings.json")
		_, err = os.Stat(settingsPath)
		assert.NoError(t, err)
	})

	t.Run("Load existing settings", func(t *testing.T) {
		// Create a settings file
		settingsPath := filepath.Join(tempDir, "settings.json")
		content := `{
			"currencies": {
				"enabled": ["USD", "GBP"],
				"default": "GBP",
				"fixed_rates": {}
			},
			"ui": {
				"date_format": "02/01/2006",
				"decimal_places": 3,
				"theme": "dark"
			},
			"version": "1.0.0"
		}`
		err := os.WriteFile(settingsPath, []byte(content), 0644)
		require.NoError(t, err)

		// Load settings
		service, err := NewSettingsService(tempDir)
		require.NoError(t, err)

		settings := service.Get()
		assert.Equal(t, []string{"USD", "GBP"}, settings.Currencies.Enabled)
		assert.Equal(t, "GBP", settings.Currencies.Default)
		assert.Equal(t, "02/01/2006", settings.UI.DateFormat)
		assert.Equal(t, 3, settings.UI.DecimalPlaces)
		assert.Equal(t, "dark", settings.UI.Theme)
	})

	t.Run("Enable and disable currencies", func(t *testing.T) {
		service, err := NewSettingsService(tempDir)
		require.NoError(t, err)

		// Enable a new currency
		err = service.EnableCurrency("JPY")
		assert.NoError(t, err)
		assert.True(t, service.IsCurrencyEnabled("JPY"))

		// Enable already enabled currency (should be idempotent)
		err = service.EnableCurrency("USD")
		assert.NoError(t, err)
		assert.True(t, service.IsCurrencyEnabled("USD"))

		// Get enabled currencies
		enabled := service.GetEnabledCurrencies()
		assert.Contains(t, enabled, "JPY")
		assert.Contains(t, enabled, "USD")
	})

	t.Run("Cannot disable default currency", func(t *testing.T) {
		service, err := NewSettingsService(tempDir)
		require.NoError(t, err)

		// Create mock transaction service
		db := test.SetupTestDB(t)
		txRepo := repository.NewTransactionRepository(db)
		currencyService := NewCurrencyService(service)
		txService := NewTransactionService(txRepo, currencyService)

		// Try to disable default currency
		err = service.DisableCurrency("USD", txService)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to disable currency USD")
		assert.True(t, service.IsCurrencyEnabled("USD"))
	})

	t.Run("Cannot disable currency with transactions", func(t *testing.T) {
		service, err := NewSettingsService(tempDir)
		require.NoError(t, err)

		// Create mock transaction service with EUR transaction
		db := test.SetupTestDB(t)
		txRepo := repository.NewTransactionRepository(db)
		categoryRepo := repository.NewCategoryRepository(db)
		currencyService := NewCurrencyService(service)
		txService := NewTransactionService(txRepo, currencyService)

		// Create a category first
		category := &models.Category{
			Name: "Test",
			Type: models.TransactionTypeExpense,
		}
		err = categoryRepo.Create(category)
		require.NoError(t, err)

		// Create EUR transaction
		tx := &models.Transaction{
			Type:        models.TransactionTypeExpense,
			Amount:      100,
			Currency:    "EUR",
			CategoryID:  category.ID,
			Description: "Test transaction",
		}
		err = txService.Create(tx)
		require.NoError(t, err)

		// Try to disable EUR
		err = service.DisableCurrency("EUR", txService)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot disable currency EUR")
		assert.True(t, service.IsCurrencyEnabled("EUR"))
	})

	t.Run("Set default currency", func(t *testing.T) {
		service, err := NewSettingsService(tempDir)
		require.NoError(t, err)

		// Set new default (must be enabled)
		err = service.SetDefaultCurrency("EUR")
		assert.NoError(t, err)
		assert.Equal(t, "EUR", service.GetDefaultCurrency())

		// Try to set non-enabled currency as default
		err = service.SetDefaultCurrency("JPY")
		assert.Error(t, err)
		assert.Equal(t, "EUR", service.GetDefaultCurrency()) // Should remain EUR
	})

	t.Run("Fixed exchange rates", func(t *testing.T) {
		service, err := NewSettingsService(tempDir)
		require.NoError(t, err)

		// Check default AED rate
		rate, exists := service.GetFixedRate("AED")
		assert.True(t, exists)
		assert.Equal(t, 3.6725, rate)

		// Set new fixed rate
		err = service.SetFixedRate("HKD", 7.85)
		assert.NoError(t, err)

		rate, exists = service.GetFixedRate("HKD")
		assert.True(t, exists)
		assert.Equal(t, 7.85, rate)

		// Remove fixed rate
		err = service.RemoveFixedRate("HKD")
		assert.NoError(t, err)

		_, exists = service.GetFixedRate("HKD")
		assert.False(t, exists)
	})

	t.Run("Concurrent access safety", func(t *testing.T) {
		service, err := NewSettingsService(tempDir)
		require.NoError(t, err)

		// Run concurrent operations
		done := make(chan bool, 3)

		go func() {
			for i := 0; i < 100; i++ {
				_ = service.GetEnabledCurrencies()
			}
			done <- true
		}()

		go func() {
			for i := 0; i < 100; i++ {
				_ = service.IsCurrencyEnabled("USD")
			}
			done <- true
		}()

		go func() {
			for i := 0; i < 100; i++ {
				_ = service.GetDefaultCurrency()
			}
			done <- true
		}()

		// Wait for all goroutines
		for i := 0; i < 3; i++ {
			<-done
		}
	})
}