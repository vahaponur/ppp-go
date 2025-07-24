package ppp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	DefaultCurrencyAPI = "https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api@latest/v1"
)

// CurrencyClient handles currency exchange rate API interactions
type CurrencyClient struct {
	baseURL string
	client  *resty.Client
}

// NewCurrencyClient creates a new currency API client
func NewCurrencyClient(baseURL string) *CurrencyClient {
	if baseURL == "" {
		baseURL = DefaultCurrencyAPI
	}
	
	return &CurrencyClient{
		baseURL: baseURL,
		client: resty.New().
			SetTimeout(10 * time.Second).
			SetRetryCount(3).
			SetRetryWaitTime(500 * time.Millisecond),
	}
}

// GetExchangeRate fetches the exchange rate between two currencies
func (c *CurrencyClient) GetExchangeRate(ctx context.Context, from, to string) (*ExchangeRate, error) {
	from = strings.ToLower(from)
	to = strings.ToLower(to)
	
	url := fmt.Sprintf("%s/currencies/%s.json", c.baseURL, from)
	
	resp, err := c.client.R().
		SetContext(ctx).
		Get(url)
	
	if err != nil {
		return nil, fmt.Errorf("failed to fetch exchange rate: %w", err)
	}
	
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode(), resp.String())
	}
	
	var data map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &data); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	// Extract the date
	dateStr, ok := data["date"].(string)
	if !ok {
		dateStr = time.Now().Format("2006-01-02")
	}
	
	// Extract rates
	rates, ok := data[from].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no rates found for currency %s", from)
	}
	
	rate, ok := rates[to].(float64)
	if !ok {
		return nil, fmt.Errorf("no exchange rate found for %s to %s", from, to)
	}
	
	lastUpdated, _ := time.Parse("2006-01-02", dateStr)
	
	return &ExchangeRate{
		From:        strings.ToUpper(from),
		To:          strings.ToUpper(to),
		Rate:        rate,
		LastUpdated: lastUpdated,
	}, nil
}

// GetUSDRates fetches all exchange rates from USD
func (c *CurrencyClient) GetUSDRates(ctx context.Context) (map[string]float64, error) {
	url := fmt.Sprintf("%s/currencies/usd.json", c.baseURL)
	
	resp, err := c.client.R().
		SetContext(ctx).
		Get(url)
	
	if err != nil {
		return nil, fmt.Errorf("failed to fetch USD rates: %w", err)
	}
	
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode(), resp.String())
	}
	
	var response CurrencyAPIResponse
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	// Convert currency codes to uppercase
	upperRates := make(map[string]float64)
	for code, rate := range response.USD {
		upperRates[strings.ToUpper(code)] = rate
	}
	
	return upperRates, nil
}

// GetAllCurrencies fetches a list of all supported currencies
func (c *CurrencyClient) GetAllCurrencies(ctx context.Context) ([]string, error) {
	// Get USD rates which contains all supported currencies
	rates, err := c.GetUSDRates(ctx)
	if err != nil {
		return nil, err
	}
	
	currencies := make([]string, 0, len(rates))
	for code := range rates {
		currencies = append(currencies, code)
	}
	
	return currencies, nil
}

// ConvertAmount converts an amount from one currency to another
func (c *CurrencyClient) ConvertAmount(ctx context.Context, amount float64, from, to string) (float64, error) {
	rate, err := c.GetExchangeRate(ctx, from, to)
	if err != nil {
		return 0, err
	}
	
	return amount * rate.Rate, nil
}

// GetHistoricalRate fetches exchange rate for a specific date
// Note: This API might not support historical data, so this is a placeholder
func (c *CurrencyClient) GetHistoricalRate(ctx context.Context, from, to string, date time.Time) (*ExchangeRate, error) {
	// For now, just return current rate
	// In a real implementation, you might use a different API that supports historical data
	return c.GetExchangeRate(ctx, from, to)
}