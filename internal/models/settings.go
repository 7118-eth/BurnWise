package models

import (
	"time"
)

// Settings represents the application configuration
type Settings struct {
	Currencies CurrencySettings `json:"currencies"`
	UI         UISettings       `json:"ui"`
	Version    string          `json:"version"`
}

// CurrencySettings holds currency-related configuration
type CurrencySettings struct {
	Enabled    []string           `json:"enabled"`
	Default    string             `json:"default"`
	FixedRates map[string]float64 `json:"fixed_rates"`
}

// UISettings holds UI-related preferences
type UISettings struct {
	DateFormat    string `json:"date_format"`
	DecimalPlaces int    `json:"decimal_places"`
	Theme         string `json:"theme"`
}

// DefaultSettings returns the default application settings
func DefaultSettings() *Settings {
	return &Settings{
		Currencies: CurrencySettings{
			Enabled: []string{"USD", "EUR", "AED"},
			Default: "USD",
			FixedRates: map[string]float64{
				"AED": 3.6725,
			},
		},
		UI: UISettings{
			DateFormat:    "2006-01-02",
			DecimalPlaces: 2,
			Theme:         "default",
		},
		Version: "1.0.0",
	}
}

// ValidateCurrency checks if a currency is enabled
func (s *Settings) ValidateCurrency(currency string) bool {
	for _, c := range s.Currencies.Enabled {
		if c == currency {
			return true
		}
	}
	return false
}

// AddCurrency enables a new currency
func (s *Settings) AddCurrency(currency string) {
	// Check if already enabled
	if s.ValidateCurrency(currency) {
		return
	}
	s.Currencies.Enabled = append(s.Currencies.Enabled, currency)
}

// RemoveCurrency disables a currency
func (s *Settings) RemoveCurrency(currency string) bool {
	// Cannot remove default currency
	if currency == s.Currencies.Default {
		return false
	}

	newEnabled := make([]string, 0, len(s.Currencies.Enabled)-1)
	removed := false
	for _, c := range s.Currencies.Enabled {
		if c != currency {
			newEnabled = append(newEnabled, c)
		} else {
			removed = true
		}
	}
	s.Currencies.Enabled = newEnabled
	return removed
}

// SetDefaultCurrency changes the default currency
func (s *Settings) SetDefaultCurrency(currency string) bool {
	// Must be an enabled currency
	if !s.ValidateCurrency(currency) {
		return false
	}
	s.Currencies.Default = currency
	return true
}

// CategoryHistory tracks changes to categories
type CategoryHistory struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	CategoryID       uint           `gorm:"not null" json:"category_id"`
	OldName          string         `json:"old_name"`
	NewName          string         `json:"new_name"`
	Action           string         `gorm:"type:varchar(20);not null" json:"action"` // rename, merge, delete
	TargetCategoryID *uint          `json:"target_category_id,omitempty"`
	ChangedAt        time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"changed_at"`
	CreatedAt        time.Time      `json:"created_at"`
}

const (
	CategoryActionRename = "rename"
	CategoryActionMerge  = "merge"
	CategoryActionDelete = "delete"
)