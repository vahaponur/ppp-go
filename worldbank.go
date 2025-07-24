package ppp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	DefaultWorldBankAPI = "https://api.worldbank.org/v2"
	PPPIndicatorCode    = "PA.NUS.PPP"
)

// WorldBankClient handles World Bank API interactions
type WorldBankClient struct {
	baseURL string
	client  *resty.Client
}

// NewWorldBankClient creates a new World Bank API client
func NewWorldBankClient(baseURL string) *WorldBankClient {
	if baseURL == "" {
		baseURL = DefaultWorldBankAPI
	}
	
	return &WorldBankClient{
		baseURL: baseURL,
		client: resty.New().
			SetTimeout(30 * time.Second).
			SetRetryCount(3).
			SetRetryWaitTime(1 * time.Second),
	}
}

// GetPPP fetches the most recent PPP data for a country
func (w *WorldBankClient) GetPPP(ctx context.Context, countryCode string) (*PPPData, error) {
	// Get data for the last 10 years to find the most recent available
	endYear := time.Now().Year()
	startYear := endYear - 10
	
	url := fmt.Sprintf("%s/country/%s/indicator/%s", w.baseURL, countryCode, PPPIndicatorCode)
	
	resp, err := w.client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"format":   "json",
			"date":     fmt.Sprintf("%d:%d", startYear, endYear),
			"per_page": "50",
		}).
		Get(url)
	
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PPP data: %w", err)
	}
	
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode(), resp.String())
	}
	
	var response WorldBankResponse
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	if len(response) < 2 {
		return nil, fmt.Errorf("invalid response format")
	}
	
	// Parse the data array
	dataBytes, err := json.Marshal(response[1])
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	
	var dataPoints []IndicatorData
	if err := json.Unmarshal(dataBytes, &dataPoints); err != nil {
		return nil, fmt.Errorf("failed to parse data points: %w", err)
	}
	
	// Find the most recent non-null value
	for _, dp := range dataPoints {
		if dp.Value != nil && *dp.Value > 0 {
			year, _ := strconv.Atoi(dp.Date)
			return &PPPData{
				CountryCode: dp.Country.ID,
				CountryName: dp.Country.Value,
				Year:        year,
				Factor:      *dp.Value,
				LastUpdated: time.Now(),
				Source:      "World Bank",
			}, nil
		}
	}
	
	return nil, fmt.Errorf("no PPP data available for country %s", countryCode)
}

// GetHistoricalPPP fetches historical PPP data for a country
func (w *WorldBankClient) GetHistoricalPPP(ctx context.Context, countryCode string, startYear, endYear int) ([]PPPData, error) {
	url := fmt.Sprintf("%s/country/%s/indicator/%s", w.baseURL, countryCode, PPPIndicatorCode)
	
	resp, err := w.client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"format":   "json",
			"date":     fmt.Sprintf("%d:%d", startYear, endYear),
			"per_page": "100",
		}).
		Get(url)
	
	if err != nil {
		return nil, fmt.Errorf("failed to fetch historical PPP data: %w", err)
	}
	
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode(), resp.String())
	}
	
	var response WorldBankResponse
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	if len(response) < 2 {
		return nil, fmt.Errorf("invalid response format")
	}
	
	dataBytes, err := json.Marshal(response[1])
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	
	var dataPoints []IndicatorData
	if err := json.Unmarshal(dataBytes, &dataPoints); err != nil {
		return nil, fmt.Errorf("failed to parse data points: %w", err)
	}
	
	var results []PPPData
	for _, dp := range dataPoints {
		if dp.Value != nil && *dp.Value > 0 {
			year, _ := strconv.Atoi(dp.Date)
			results = append(results, PPPData{
				CountryCode: dp.Country.ID,
				CountryName: dp.Country.Value,
				Year:        year,
				Factor:      *dp.Value,
				LastUpdated: time.Now(),
				Source:      "World Bank",
			})
		}
	}
	
	return results, nil
}

// GetCountries fetches all available countries
func (w *WorldBankClient) GetCountries(ctx context.Context) ([]Country, error) {
	url := fmt.Sprintf("%s/country", w.baseURL)
	
	resp, err := w.client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"format":   "json",
			"per_page": "300",
		}).
		Get(url)
	
	if err != nil {
		return nil, fmt.Errorf("failed to fetch countries: %w", err)
	}
	
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode(), resp.String())
	}
	
	var response WorldBankResponse
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	if len(response) < 2 {
		return nil, fmt.Errorf("invalid response format")
	}
	
	dataBytes, err := json.Marshal(response[1])
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	
	var countries []Country
	if err := json.Unmarshal(dataBytes, &countries); err != nil {
		return nil, fmt.Errorf("failed to parse countries: %w", err)
	}
	
	// Filter out aggregates (regions, income groups, etc.)
	var result []Country
	for _, c := range countries {
		// Only include actual countries (they have capital cities)
		if c.CapitalCity != "" {
			result = append(result, c)
		}
	}
	
	return result, nil
}

// SearchIndicators searches for indicators by keyword
func (w *WorldBankClient) SearchIndicators(ctx context.Context, search string) ([]Indicator, error) {
	apiURL := fmt.Sprintf("%s/indicator", w.baseURL)
	
	resp, err := w.client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"format":   "json",
			"per_page": "100",
			"source":   "2", // World Development Indicators
			"search":   url.QueryEscape(search),
		}).
		Get(apiURL)
	
	if err != nil {
		return nil, fmt.Errorf("failed to search indicators: %w", err)
	}
	
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode(), resp.String())
	}
	
	var response WorldBankResponse
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	if len(response) < 2 {
		return nil, fmt.Errorf("invalid response format")
	}
	
	dataBytes, err := json.Marshal(response[1])
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	
	var indicators []Indicator
	if err := json.Unmarshal(dataBytes, &indicators); err != nil {
		return nil, fmt.Errorf("failed to parse indicators: %w", err)
	}
	
	return indicators, nil
}

// GetIndicatorData fetches data for any indicator
func (w *WorldBankClient) GetIndicatorData(ctx context.Context, countryCode, indicatorCode string, startYear, endYear int) ([]IndicatorData, error) {
	url := fmt.Sprintf("%s/country/%s/indicator/%s", w.baseURL, countryCode, indicatorCode)
	
	resp, err := w.client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"format":   "json",
			"date":     fmt.Sprintf("%d:%d", startYear, endYear),
			"per_page": "100",
		}).
		Get(url)
	
	if err != nil {
		return nil, fmt.Errorf("failed to fetch indicator data: %w", err)
	}
	
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode(), resp.String())
	}
	
	var response WorldBankResponse
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	if len(response) < 2 {
		return nil, fmt.Errorf("invalid response format")
	}
	
	dataBytes, err := json.Marshal(response[1])
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	
	var dataPoints []IndicatorData
	if err := json.Unmarshal(dataBytes, &dataPoints); err != nil {
		return nil, fmt.Errorf("failed to parse data points: %w", err)
	}
	
	return dataPoints, nil
}