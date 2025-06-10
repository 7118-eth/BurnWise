package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCurrencyService_FixedRate(t *testing.T) {
	service := NewCurrencyService()
	
	// Test AED to USD (fixed rate)
	usdAmount, err := service.ConvertToUSD(100.00, "AED")
	require.NoError(t, err)
	assert.InDelta(t, 27.23, usdAmount, 0.01)
	
	// Test USD to AED
	aedAmount, err := service.ConvertFromUSD(100.00, "AED")
	require.NoError(t, err)
	assert.InDelta(t, 367.25, aedAmount, 0.01)
}

func TestCurrencyService_USDConversion(t *testing.T) {
	service := NewCurrencyService()
	
	// Test USD to USD (should return same amount)
	amount, err := service.ConvertToUSD(100.00, "USD")
	require.NoError(t, err)
	assert.Equal(t, 100.00, amount)
	
	amount, err = service.ConvertFromUSD(100.00, "USD")
	require.NoError(t, err)
	assert.Equal(t, 100.00, amount)
}

func TestCurrencyService_SupportedCurrencies(t *testing.T) {
	service := NewCurrencyService()
	
	supported := service.GetSupportedCurrencies()
	assert.Contains(t, supported, "USD")
	assert.Contains(t, supported, "EUR")
	assert.Contains(t, supported, "AED")
	
	assert.True(t, service.IsSupported("USD"))
	assert.True(t, service.IsSupported("AED"))
	assert.False(t, service.IsSupported("XXX"))
}