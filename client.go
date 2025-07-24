package ppp

import (
	"context"
	"fmt"
	"time"
)

// Client is the main PPP client
type Client struct {
	worldBank     *WorldBankClient
	currency      *CurrencyClient
	cache         *Cache
	cacheEnabled  bool
	cacheDuration time.Duration
	timeout       time.Duration
}

// Option is a functional option for configuring the client
type Option func(*Client)

// WithCache enables caching with the specified duration
func WithCache(duration time.Duration) Option {
	return func(c *Client) {
		c.cacheEnabled = true
		c.cacheDuration = duration
		c.cache = NewCache(duration, duration*2)
	}
}

// WithoutCache disables caching
func WithoutCache() Option {
	return func(c *Client) {
		c.cacheEnabled = false
		c.cache = nil
	}
}

// WithTimeout sets the client timeout
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithWorldBankURL sets a custom World Bank API URL
func WithWorldBankURL(url string) Option {
	return func(c *Client) {
		c.worldBank = NewWorldBankClient(url)
	}
}

// WithCurrencyURL sets a custom currency API URL
func WithCurrencyURL(url string) Option {
	return func(c *Client) {
		c.currency = NewCurrencyClient(url)
	}
}

// NewClient creates a new PPP client with options
func NewClient(opts ...Option) *Client {
	// Default client with cache enabled
	client := &Client{
		worldBank:     NewWorldBankClient(""),
		currency:      NewCurrencyClient(""),
		cacheEnabled:  true,
		cacheDuration: 24 * time.Hour,
		timeout:       30 * time.Second,
	}
	
	// Enable cache by default
	client.cache = NewCache(client.cacheDuration, client.cacheDuration*2)
	
	// Apply options
	for _, opt := range opts {
		opt(client)
	}
	
	return client
}

// GetPPP fetches PPP data for a country
func (c *Client) GetPPP(ctx context.Context, countryCode string) (*PPPData, error) {
	// Check cache first if enabled
	if c.cacheEnabled && c.cache != nil {
		if ppp, found := c.cache.GetPPP(countryCode); found {
			return ppp, nil
		}
	}
	
	// Fetch from API
	ppp, err := c.worldBank.GetPPP(ctx, countryCode)
	if err != nil {
		return nil, err
	}
	
	// Store in cache if enabled
	if c.cacheEnabled && c.cache != nil {
		c.cache.SetPPP(countryCode, ppp, c.cacheDuration)
	}
	
	return ppp, nil
}

// GetExchangeRate fetches exchange rate between two currencies
func (c *Client) GetExchangeRate(ctx context.Context, from, to string) (*ExchangeRate, error) {
	// Check cache first if enabled
	if c.cacheEnabled && c.cache != nil {
		if rate, found := c.cache.GetExchangeRate(from, to); found {
			return rate, nil
		}
	}
	
	// Fetch from API
	rate, err := c.currency.GetExchangeRate(ctx, from, to)
	if err != nil {
		return nil, err
	}
	
	// Store in cache if enabled
	if c.cacheEnabled && c.cache != nil {
		// Exchange rates cache for shorter duration (1 hour)
		c.cache.SetExchangeRate(from, to, rate, time.Hour)
	}
	
	return rate, nil
}

// Recommend calculates recommended price based on PPP
func (c *Client) Recommend(ctx context.Context, price float64, fromCurrency, toCountry string) (*PriceRecommendation, error) {
	// Get PPP data
	ppp, err := c.GetPPP(ctx, toCountry)
	if err != nil {
		return nil, fmt.Errorf("failed to get PPP data: %w", err)
	}
	
	// Get currency mapping for the country
	toCurrency := c.getCurrencyForCountry(toCountry)
	
	// Get exchange rate
	rate, err := c.GetExchangeRate(ctx, fromCurrency, toCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange rate: %w", err)
	}
	
	// Calculate recommended price
	// PPP factor is LCU per international $ (e.g., 11.55 TRY per 1 international $)
	// So for $100, the PPP-adjusted price is 100 * 11.55 = 1155 TRY
	recommendedPrice := price * ppp.Factor
	
	// Calculate discount percentage compared to market exchange rate
	// Market price would be: $100 * 40.47 = 4047 TRY
	// PPP price is: $100 * 11.55 = 1155 TRY
	// So the discount is about 71%
	normalPrice := price * rate.Rate
	discountPercentage := ((normalPrice - recommendedPrice) / normalPrice) * 100
	
	return &PriceRecommendation{
		OriginalPrice:      price,
		OriginalCurrency:   fromCurrency,
		RecommendedPrice:   recommendedPrice,
		TargetCurrency:     toCurrency,
		PPPFactor:          ppp.Factor,
		ExchangeRate:       rate.Rate,
		DiscountPercentage: discountPercentage,
	}, nil
}

// GetCountries fetches all available countries
func (c *Client) GetCountries(ctx context.Context) ([]Country, error) {
	// Check cache first if enabled
	if c.cacheEnabled && c.cache != nil {
		if countries, found := c.cache.GetCountries(); found {
			return countries, nil
		}
	}
	
	// Fetch from API
	countries, err := c.worldBank.GetCountries(ctx)
	if err != nil {
		return nil, err
	}
	
	// Store in cache if enabled
	if c.cacheEnabled && c.cache != nil {
		// Countries don't change often, cache for longer
		c.cache.SetCountries(countries, 7*24*time.Hour)
	}
	
	return countries, nil
}

// SearchIndicators searches for indicators by keyword
func (c *Client) SearchIndicators(ctx context.Context, search string) ([]Indicator, error) {
	// Check cache first if enabled
	if c.cacheEnabled && c.cache != nil {
		if indicators, found := c.cache.GetIndicators(search); found {
			return indicators, nil
		}
	}
	
	// Fetch from API
	indicators, err := c.worldBank.SearchIndicators(ctx, search)
	if err != nil {
		return nil, err
	}
	
	// Store in cache if enabled
	if c.cacheEnabled && c.cache != nil {
		c.cache.SetIndicators(search, indicators, 24*time.Hour)
	}
	
	return indicators, nil
}

// GetHistoricalPPP fetches historical PPP data
func (c *Client) GetHistoricalPPP(ctx context.Context, countryCode string, startYear, endYear int) ([]PPPData, error) {
	return c.worldBank.GetHistoricalPPP(ctx, countryCode, startYear, endYear)
}

// GetIndicatorData fetches data for any indicator
func (c *Client) GetIndicatorData(ctx context.Context, countryCode, indicatorCode string, startYear, endYear int) ([]IndicatorData, error) {
	return c.worldBank.GetIndicatorData(ctx, countryCode, indicatorCode, startYear, endYear)
}

// AnalyzePPPTrend analyzes PPP trend for a country
func (c *Client) AnalyzePPPTrend(ctx context.Context, countryCode string, startYear, endYear int) (*PPPTrendAnalysis, error) {
	data, err := c.GetHistoricalPPP(ctx, countryCode, startYear, endYear)
	if err != nil {
		return nil, err
	}
	
	if len(data) == 0 {
		return nil, fmt.Errorf("no data available for analysis")
	}
	
	// Calculate average
	var sum float64
	for _, d := range data {
		sum += d.Factor
	}
	average := sum / float64(len(data))
	
	// Determine trend
	trend := "stable"
	if len(data) >= 2 {
		firstYear := data[len(data)-1].Factor // Oldest
		lastYear := data[0].Factor            // Newest
		change := ((lastYear - firstYear) / firstYear) * 100
		
		if change > 10 {
			trend = "increasing"
		} else if change < -10 {
			trend = "decreasing"
		}
	}
	
	// Calculate volatility (standard deviation)
	var variance float64
	for _, d := range data {
		variance += (d.Factor - average) * (d.Factor - average)
	}
	volatility := 0.0
	if len(data) > 1 {
		volatility = variance / float64(len(data)-1)
	}
	
	return &PPPTrendAnalysis{
		Country:    countryCode,
		StartYear:  startYear,
		EndYear:    endYear,
		Average:    average,
		Trend:      trend,
		Volatility: volatility,
		DataPoints: len(data),
	}, nil
}

// ComparePPP compares PPP factors across multiple countries
func (c *Client) ComparePPP(ctx context.Context, countryCodes []string) ([]CountryComparison, error) {
	comparisons := make([]CountryComparison, 0, len(countryCodes))
	
	for i, code := range countryCodes {
		ppp, err := c.GetPPP(ctx, code)
		if err != nil {
			continue // Skip countries with errors
		}
		
		comp := CountryComparison{
			Country:     code,
			CountryName: ppp.CountryName,
			Factor:      ppp.Factor,
			PercentOfUS: (1.0 / ppp.Factor) * 100, // Assuming US PPP = 1
			Rank:        i + 1,
		}
		comparisons = append(comparisons, comp)
	}
	
	// Sort by PPP factor (lower factor = higher purchasing power)
	for i := 0; i < len(comparisons); i++ {
		for j := i + 1; j < len(comparisons); j++ {
			if comparisons[i].Factor > comparisons[j].Factor {
				comparisons[i], comparisons[j] = comparisons[j], comparisons[i]
			}
		}
	}
	
	// Update ranks
	for i := range comparisons {
		comparisons[i].Rank = i + 1
	}
	
	return comparisons, nil
}

// ExportCache exports cache to file
func (c *Client) ExportCache(filename string) error {
	if !c.cacheEnabled || c.cache == nil {
		return fmt.Errorf("cache is disabled")
	}
	return c.cache.ExportToFile(filename)
}

// ImportCache imports cache from file
func (c *Client) ImportCache(filename string) error {
	if !c.cacheEnabled || c.cache == nil {
		return fmt.Errorf("cache is disabled")
	}
	return c.cache.ImportFromFile(filename)
}

// ClearCache clears all cached data
func (c *Client) ClearCache() {
	if c.cacheEnabled && c.cache != nil {
		c.cache.Clear()
	}
}

// getCurrencyForCountry maps country code to currency code
// This is a simplified mapping - in production you'd want a complete list
func (c *Client) getCurrencyForCountry(countryCode string) string {
	currencyMap := map[string]string{
		"US": "USD",
		"TR": "TRY",
		"DE": "EUR",
		"FR": "EUR",
		"IT": "EUR",
		"ES": "EUR",
		"GB": "GBP",
		"JP": "JPY",
		"CN": "CNY",
		"IN": "INR",
		"BR": "BRL",
		"RU": "RUB",
		"CA": "CAD",
		"AU": "AUD",
		"MX": "MXN",
		"KR": "KRW",
		"ID": "IDR",
		"SA": "SAR",
		"AR": "ARS",
		"ZA": "ZAR",
		"NG": "NGN",
		"EG": "EGP",
		"PK": "PKR",
		"BD": "BDT",
		"VN": "VND",
		"TH": "THB",
		"MY": "MYR",
		"SG": "SGD",
		"PH": "PHP",
		"NZ": "NZD",
		"CH": "CHF",
		"SE": "SEK",
		"NO": "NOK",
		"DK": "DKK",
		"PL": "PLN",
		"CZ": "CZK",
		"HU": "HUF",
		"RO": "RON",
		"UA": "UAH",
		"IL": "ILS",
		"AE": "AED",
		"CL": "CLP",
		"CO": "COP",
		"PE": "PEN",
	}
	
	if currency, ok := currencyMap[countryCode]; ok {
		return currency
	}
	
	// Default to USD if not found
	return "USD"
}