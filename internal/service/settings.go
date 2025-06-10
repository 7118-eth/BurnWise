package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"budget-tracker/internal/models"
)

// SettingsService manages application settings
type SettingsService struct {
	settings     *models.Settings
	settingsPath string
	mu           sync.RWMutex
}

// NewSettingsService creates a new settings service
func NewSettingsService(dataDir string) (*SettingsService, error) {
	settingsPath := filepath.Join(dataDir, "settings.json")
	s := &SettingsService{
		settingsPath: settingsPath,
	}

	// Load existing settings or create default
	if err := s.Load(); err != nil {
		// If file doesn't exist, create default settings
		if os.IsNotExist(err) {
			s.settings = models.DefaultSettings()
			if err := s.Save(); err != nil {
				return nil, fmt.Errorf("failed to save default settings: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to load settings: %w", err)
		}
	}

	return s, nil
}

// Load reads settings from file
func (s *SettingsService) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.settingsPath)
	if err != nil {
		return err
	}

	var settings models.Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	s.settings = &settings
	return nil
}

// Save writes settings to file
func (s *SettingsService) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s.settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Write to temp file first
	tempPath := s.settingsPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings: %w", err)
	}

	// Rename to actual file (atomic operation)
	if err := os.Rename(tempPath, s.settingsPath); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("failed to save settings: %w", err)
	}

	return nil
}

// Get returns a copy of current settings
func (s *SettingsService) Get() models.Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.settings
}

// Update applies changes to settings
func (s *SettingsService) Update(fn func(*models.Settings) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Apply changes
	if err := fn(s.settings); err != nil {
		return err
	}

	// Save to file
	return s.Save()
}

// GetEnabledCurrencies returns list of enabled currencies
func (s *SettingsService) GetEnabledCurrencies() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Return a copy to prevent external modification
	currencies := make([]string, len(s.settings.Currencies.Enabled))
	copy(currencies, s.settings.Currencies.Enabled)
	return currencies
}

// GetDefaultCurrency returns the default currency
func (s *SettingsService) GetDefaultCurrency() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings.Currencies.Default
}

// IsCurrencyEnabled checks if a currency is enabled
func (s *SettingsService) IsCurrencyEnabled(currency string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings.ValidateCurrency(currency)
}

// EnableCurrency adds a currency to the enabled list
func (s *SettingsService) EnableCurrency(currency string) error {
	return s.Update(func(settings *models.Settings) error {
		settings.AddCurrency(currency)
		return nil
	})
}

// DisableCurrency removes a currency from the enabled list
func (s *SettingsService) DisableCurrency(currency string, transactionService *TransactionService) error {
	// First check if any transactions use this currency
	count, err := transactionService.CountByCurrency(currency)
	if err != nil {
		return fmt.Errorf("failed to check currency usage: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("cannot disable currency %s: %d transactions use this currency", currency, count)
	}

	return s.Update(func(settings *models.Settings) error {
		if !settings.RemoveCurrency(currency) {
			return fmt.Errorf("failed to disable currency %s", currency)
		}
		return nil
	})
}

// SetDefaultCurrency changes the default currency
func (s *SettingsService) SetDefaultCurrency(currency string) error {
	return s.Update(func(settings *models.Settings) error {
		if !settings.SetDefaultCurrency(currency) {
			return fmt.Errorf("invalid currency: %s", currency)
		}
		return nil
	})
}

// GetFixedRate returns the fixed exchange rate for a currency if it exists
func (s *SettingsService) GetFixedRate(currency string) (float64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	rate, exists := s.settings.Currencies.FixedRates[currency]
	return rate, exists
}

// SetFixedRate sets a fixed exchange rate for a currency
func (s *SettingsService) SetFixedRate(currency string, rate float64) error {
	return s.Update(func(settings *models.Settings) error {
		if settings.Currencies.FixedRates == nil {
			settings.Currencies.FixedRates = make(map[string]float64)
		}
		settings.Currencies.FixedRates[currency] = rate
		return nil
	})
}

// RemoveFixedRate removes a fixed exchange rate for a currency
func (s *SettingsService) RemoveFixedRate(currency string) error {
	return s.Update(func(settings *models.Settings) error {
		delete(settings.Currencies.FixedRates, currency)
		return nil
	})
}