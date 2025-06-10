package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

var fixedRates = map[string]float64{
	"AED": 3.6725,
}

type exchangeRateResponse struct {
	Result             string             `json:"result"`
	ConversionRates    map[string]float64 `json:"conversion_rates"`
	TimeLastUpdateUnix int64              `json:"time_last_update_unix"`
}

type CurrencyService struct {
	cache     map[string]*rateCache
	cacheMutex sync.RWMutex
	apiKey    string
}

type rateCache struct {
	rate      float64
	timestamp time.Time
}

func NewCurrencyService() *CurrencyService {
	return &CurrencyService{
		cache:  make(map[string]*rateCache),
		apiKey: "free", // Using free tier
	}
}

func (s *CurrencyService) ConvertToUSD(amount float64, currency string) (float64, error) {
	if currency == "USD" {
		return amount, nil
	}

	rate, err := s.GetExchangeRate(currency)
	if err != nil {
		return 0, err
	}

	return amount / rate, nil
}

func (s *CurrencyService) ConvertFromUSD(amount float64, currency string) (float64, error) {
	if currency == "USD" {
		return amount, nil
	}

	rate, err := s.GetExchangeRate(currency)
	if err != nil {
		return 0, err
	}

	return amount * rate, nil
}

func (s *CurrencyService) GetExchangeRate(currency string) (float64, error) {
	if rate, ok := fixedRates[currency]; ok {
		return rate, nil
	}

	s.cacheMutex.RLock()
	if cached, ok := s.cache[currency]; ok {
		if time.Since(cached.timestamp) < time.Hour {
			s.cacheMutex.RUnlock()
			return cached.rate, nil
		}
	}
	s.cacheMutex.RUnlock()

	rate, err := s.fetchExchangeRate(currency)
	if err != nil {
		return 0, err
	}

	s.cacheMutex.Lock()
	s.cache[currency] = &rateCache{
		rate:      rate,
		timestamp: time.Now(),
	}
	s.cacheMutex.Unlock()

	return rate, nil
}

func (s *CurrencyService) fetchExchangeRate(currency string) (float64, error) {
	url := fmt.Sprintf("https://api.exchangerate-api.com/v4/latest/USD")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch exchange rate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var data struct {
		Rates map[string]float64 `json:"rates"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	rate, ok := data.Rates[currency]
	if !ok {
		return 0, fmt.Errorf("currency %s not supported", currency)
	}

	return rate, nil
}

func (s *CurrencyService) GetSupportedCurrencies() []string {
	return []string{
		"USD", "EUR", "GBP", "JPY", "CHF", "CAD", "AUD", "NZD", "AED",
	}
}

func (s *CurrencyService) IsSupported(currency string) bool {
	supported := s.GetSupportedCurrencies()
	for _, c := range supported {
		if c == currency {
			return true
		}
	}
	return false
}