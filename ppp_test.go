package ppp

import (
	"testing"
	"time"
)

func TestValidateCountryCode(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{"Valid code", "TR", false},
		{"Valid code US", "US", false},
		{"Too short", "T", true},
		{"Too long", "TUR", true},
		{"Lowercase", "tr", true},
		{"Numbers", "T1", true},
		{"Empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCountryCode(tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCountryCode(%q) error = %v, wantErr %v", tt.code, err, tt.wantErr)
			}
		})
	}
}

func TestValidateCurrencyCode(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{"Valid USD", "USD", false},
		{"Valid TRY", "TRY", false},
		{"Too short", "US", true},
		{"Too long", "USDD", true},
		{"Lowercase", "usd", true},
		{"Numbers", "US1", true},
		{"Empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCurrencyCode(tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCurrencyCode(%q) error = %v, wantErr %v", tt.code, err, tt.wantErr)
			}
		})
	}
}

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name    string
		amount  float64
		wantErr bool
	}{
		{"Valid amount", 100.50, false},
		{"Small amount", 0.01, false},
		{"Zero amount", 0, true},
		{"Negative amount", -10, true},
		{"Very large amount", 1e16, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAmount(tt.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAmount(%v) error = %v, wantErr %v", tt.amount, err, tt.wantErr)
			}
		})
	}
}

func TestConvertPrice(t *testing.T) {
	tests := []struct {
		name      string
		price     float64
		pppFactor float64
		want      float64
		wantErr   bool
	}{
		{"Normal conversion", 100, 0.5, 50, false},
		{"Factor 1", 100, 1, 100, false},
		{"Small factor", 100, 0.1, 10, false},
		{"Invalid price", -100, 0.5, 0, true},
		{"Invalid factor", 100, -0.5, 0, true},
		{"Zero factor", 100, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertPrice(tt.price, tt.pppFactor)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertPrice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ConvertPrice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoundPrice(t *testing.T) {
	tests := []struct {
		name     string
		price    float64
		currency string
		want     float64
	}{
		{"USD with decimals", 99.996, "USD", 100.00},
		{"EUR with decimals", 49.994, "EUR", 49.99},
		{"JPY no decimals", 1234.56, "JPY", 1235},
		{"KRW no decimals", 50000.99, "KRW", 50001},
		{"TRY with decimals", 123.456, "TRY", 123.46},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RoundPrice(tt.price, tt.currency)
			if got != tt.want {
				t.Errorf("RoundPrice(%v, %v) = %v, want %v", tt.price, tt.currency, got, tt.want)
			}
		})
	}
}

func TestFormatPrice(t *testing.T) {
	tests := []struct {
		name     string
		price    float64
		currency string
		want     string
	}{
		{"USD format", 99.99, "USD", "$99.99"},
		{"EUR format", 49.99, "EUR", "49.99 €"},
		{"GBP format", 75.50, "GBP", "£75.50"},
		{"TRY format", 1234.56, "TRY", "₺1234.56"},
		{"JPY format", 1000, "JPY", "¥1000"},
		{"Unknown currency", 100, "XXX", "XXX 100.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatPrice(tt.price, tt.currency)
			if got != tt.want {
				t.Errorf("FormatPrice(%v, %v) = %v, want %v", tt.price, tt.currency, got, tt.want)
			}
		})
	}
}

func TestCache(t *testing.T) {
	cache := NewCache(1*time.Minute, 2*time.Minute)

	// Test PPP cache
	pppData := &PPPData{
		CountryCode: "TR",
		CountryName: "Turkey",
		Factor:      7.5,
		Year:        2023,
	}

	cache.SetPPP("TR", pppData, 1*time.Minute)
	
	got, found := cache.GetPPP("TR")
	if !found {
		t.Error("Expected to find PPP data in cache")
	}
	if got.Factor != pppData.Factor {
		t.Errorf("Got factor %v, want %v", got.Factor, pppData.Factor)
	}

	// Test exchange rate cache
	rate := &ExchangeRate{
		From: "USD",
		To:   "TRY",
		Rate: 32.5,
	}

	cache.SetExchangeRate("USD", "TRY", rate, 1*time.Minute)
	
	gotRate, found := cache.GetExchangeRate("USD", "TRY")
	if !found {
		t.Error("Expected to find exchange rate in cache")
	}
	if gotRate.Rate != rate.Rate {
		t.Errorf("Got rate %v, want %v", gotRate.Rate, rate.Rate)
	}

	// Test cache clear
	cache.Clear()
	_, found = cache.GetPPP("TR")
	if found {
		t.Error("Expected cache to be empty after clear")
	}
}

func TestClientOptions(t *testing.T) {
	// Test default client
	client1 := NewClient()
	if !client1.cacheEnabled {
		t.Error("Expected cache to be enabled by default")
	}

	// Test without cache
	client2 := NewClient(WithoutCache())
	if client2.cacheEnabled {
		t.Error("Expected cache to be disabled")
	}

	// Test with custom timeout
	client3 := NewClient(WithTimeout(5 * time.Second))
	if client3.timeout != 5*time.Second {
		t.Errorf("Expected timeout to be 5s, got %v", client3.timeout)
	}

	// Test with custom URLs
	customWBURL := "https://custom.worldbank.org"
	client4 := NewClient(WithWorldBankURL(customWBURL))
	if client4.worldBank.baseURL != customWBURL {
		t.Errorf("Expected World Bank URL to be %s, got %s", customWBURL, client4.worldBank.baseURL)
	}
}

func TestGetCurrencyForCountry(t *testing.T) {
	client := NewClient()
	
	tests := []struct {
		country string
		want    string
	}{
		{"US", "USD"},
		{"TR", "TRY"},
		{"DE", "EUR"},
		{"GB", "GBP"},
		{"JP", "JPY"},
		{"XX", "USD"}, // Unknown country should default to USD
	}

	for _, tt := range tests {
		t.Run(tt.country, func(t *testing.T) {
			got := client.getCurrencyForCountry(tt.country)
			if got != tt.want {
				t.Errorf("getCurrencyForCountry(%v) = %v, want %v", tt.country, got, tt.want)
			}
		})
	}
}

func TestPPPError(t *testing.T) {
	err := NewPPPError(ErrCodeNoData, "test error", ErrNoData).
		WithContext("country", "TR").
		WithContext("year", 2023)

	if err.Code != ErrCodeNoData {
		t.Errorf("Expected error code %s, got %s", ErrCodeNoData, err.Code)
	}

	if err.Context["country"] != "TR" {
		t.Errorf("Expected context country to be TR, got %v", err.Context["country"])
	}

	// Test error message
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Expected error message to not be empty")
	}

	// Test unwrap
	if err.Unwrap() != ErrNoData {
		t.Error("Expected unwrapped error to be ErrNoData")
	}

	// Test error type checking
	if !IsNoDataError(err) {
		t.Error("Expected IsNoDataError to return true")
	}
}