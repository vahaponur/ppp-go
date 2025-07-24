package ppp

import (
	"time"
)

// PPPData represents purchasing power parity data for a country
type PPPData struct {
	CountryCode      string    `json:"country_code"`
	CountryName      string    `json:"country_name"`
	Year             int       `json:"year"`
	Factor           float64   `json:"factor"`
	LastUpdated      time.Time `json:"last_updated"`
	Source           string    `json:"source"`
}

// ExchangeRate represents currency exchange rate data
type ExchangeRate struct {
	From         string    `json:"from"`
	To           string    `json:"to"`
	Rate         float64   `json:"rate"`
	LastUpdated  time.Time `json:"last_updated"`
}

// PriceRecommendation represents a recommended price based on PPP
type PriceRecommendation struct {
	OriginalPrice      float64 `json:"original_price"`
	OriginalCurrency   string  `json:"original_currency"`
	RecommendedPrice   float64 `json:"recommended_price"`
	TargetCurrency     string  `json:"target_currency"`
	PPPFactor          float64 `json:"ppp_factor"`
	ExchangeRate       float64 `json:"exchange_rate"`
	DiscountPercentage float64 `json:"discount_percentage"`
}

// Country represents World Bank country data
type Country struct {
	ID           string  `json:"id"`
	ISO2Code     string  `json:"iso2Code"`
	Name         string  `json:"name"`
	Region       Region  `json:"region"`
	AdminRegion  Region  `json:"adminregion"`
	IncomeLevel  Level   `json:"incomeLevel"`
	LendingType  Level   `json:"lendingType"`
	CapitalCity  string  `json:"capitalCity"`
	Longitude    string `json:"longitude"`
	Latitude     string `json:"latitude"`
}

// Region represents a World Bank region
type Region struct {
	ID       string `json:"id"`
	ISO2Code string `json:"iso2code"`
	Value    string `json:"value"`
}

// Level represents income or lending level
type Level struct {
	ID       string `json:"id"`
	ISO2Code string `json:"iso2code"`
	Value    string `json:"value"`
}

// Indicator represents a World Bank indicator
type Indicator struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Unit        string          `json:"unit"`
	Source      IndicatorSource `json:"source"`
	SourceNote  string          `json:"sourceNote"`
	SourceOrg   string          `json:"sourceOrganization"`
	Topics      []Topic         `json:"topics"`
}

// IndicatorSource represents the source of an indicator
type IndicatorSource struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

// Topic represents a World Bank topic
type Topic struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

// IndicatorData represents data for a specific indicator
type IndicatorData struct {
	Indicator        IndicatorInfo `json:"indicator"`
	Country          CountryInfo   `json:"country"`
	CountryISO3Code  string        `json:"countryiso3code"`
	Date             string        `json:"date"`
	Value            *float64      `json:"value"`
	Unit             string        `json:"unit"`
	ObsStatus        string        `json:"obs_status"`
	Decimal          int           `json:"decimal"`
}

// IndicatorInfo contains basic indicator information
type IndicatorInfo struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

// CountryInfo contains basic country information
type CountryInfo struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

// WorldBankResponse represents the response structure from World Bank API
type WorldBankResponse []interface{}

// CurrencyAPIResponse represents the response from the currency API
type CurrencyAPIResponse struct {
	Date string                 `json:"date"`
	USD  map[string]float64     `json:"usd"`
}

// PPPTrendAnalysis represents trend analysis for PPP data
type PPPTrendAnalysis struct {
	Country      string  `json:"country"`
	StartYear    int     `json:"start_year"`
	EndYear      int     `json:"end_year"`
	Average      float64 `json:"average"`
	Trend        string  `json:"trend"` // "increasing", "decreasing", "stable"
	Volatility   float64 `json:"volatility"`
	DataPoints   int     `json:"data_points"`
}

// CountryComparison represents PPP comparison between countries
type CountryComparison struct {
	Country      string  `json:"country"`
	CountryName  string  `json:"country_name"`
	Factor       float64 `json:"factor"`
	PercentOfUS  float64 `json:"percent_of_us"`
	Rank         int     `json:"rank"`
}