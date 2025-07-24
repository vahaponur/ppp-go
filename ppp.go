package ppp

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Global default client for simple usage
var defaultClient *Client

func init() {
	// Initialize with default settings
	defaultClient = NewClient()
}

// SetDefaultClient sets the default client for package-level functions
func SetDefaultClient(client *Client) {
	defaultClient = client
}

// RecommendPrice returns recommended price based on PPP
// Returns (recommendedPrice, error)
func RecommendPrice(price float64, fromCurrency, toCountry string) (float64, error) {
	// Validate inputs
	if err := ValidateAmount(price); err != nil {
		return 0, err
	}
	
	if err := ValidateCurrencyCode(fromCurrency); err != nil {
		return 0, err
	}
	
	if err := ValidateCountryCode(toCountry); err != nil {
		return 0, err
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	rec, err := defaultClient.Recommend(ctx, price, fromCurrency, toCountry)
	if err != nil {
		return 0, NewPPPError(
			ErrCodeAPIError,
			"failed to calculate recommended price",
			err,
		).WithContext("price", price).
			WithContext("from_currency", fromCurrency).
			WithContext("to_country", toCountry)
	}
	
	return rec.RecommendedPrice, nil
}

// GetFactor returns the PPP factor for a country
// Returns (factor, error)
func GetFactor(countryCode string) (float64, error) {
	if err := ValidateCountryCode(countryCode); err != nil {
		return 0, err
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	pppData, err := defaultClient.GetPPP(ctx, countryCode)
	if err != nil {
		return 0, NewPPPError(
			ErrCodeAPIError,
			"failed to get PPP factor",
			err,
		).WithContext("country_code", countryCode)
	}
	
	return pppData.Factor, nil
}

// ConvertPrice converts a price using PPP factor
// Returns (convertedPrice, error)
func ConvertPrice(price float64, pppFactor float64) (float64, error) {
	if err := ValidateAmount(price); err != nil {
		return 0, err
	}
	
	if pppFactor <= 0 {
		return 0, NewPPPError(
			ErrCodeInvalidInput,
			"PPP factor must be positive",
			nil,
		).WithContext("ppp_factor", pppFactor)
	}
	
	return price * pppFactor, nil
}

// GetRate returns exchange rate between two currencies
// Returns (rate, error)
func GetRate(from, to string) (float64, error) {
	if err := ValidateCurrencyCode(from); err != nil {
		return 0, err
	}
	
	if err := ValidateCurrencyCode(to); err != nil {
		return 0, err
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	rate, err := defaultClient.GetExchangeRate(ctx, from, to)
	if err != nil {
		return 0, NewPPPError(
			ErrCodeAPIError,
			"failed to get exchange rate",
			err,
		).WithContext("from", from).WithContext("to", to)
	}
	
	return rate.Rate, nil
}

// QuickRecommend returns a complete price recommendation
// Returns (recommendation, error)
func QuickRecommend(price float64, fromCurrency, toCountry string) (*PriceRecommendation, error) {
	// Validate inputs
	if err := ValidateAmount(price); err != nil {
		return nil, err
	}
	
	if err := ValidateCurrencyCode(fromCurrency); err != nil {
		return nil, err
	}
	
	if err := ValidateCountryCode(toCountry); err != nil {
		return nil, err
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	rec, err := defaultClient.Recommend(ctx, price, fromCurrency, toCountry)
	if err != nil {
		return nil, NewPPPError(
			ErrCodeAPIError,
			"failed to get recommendation",
			err,
		).WithContext("price", price).
			WithContext("from_currency", fromCurrency).
			WithContext("to_country", toCountry)
	}
	
	return rec, nil
}

// ListCountries returns a list of all available countries
// Returns (countries, error)
func ListCountries() ([]Country, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	countries, err := defaultClient.GetCountries(ctx)
	if err != nil {
		return nil, NewPPPError(
			ErrCodeAPIError,
			"failed to list countries",
			err,
		)
	}
	
	return countries, nil
}

// FindIndicators searches for indicators by keyword
// Returns (indicators, error)
func FindIndicators(search string) ([]Indicator, error) {
	if search == "" {
		return nil, NewPPPError(
			ErrCodeInvalidInput,
			"search query cannot be empty",
			nil,
		)
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	indicators, err := defaultClient.SearchIndicators(ctx, search)
	if err != nil {
		return nil, NewPPPError(
			ErrCodeAPIError,
			"failed to search indicators",
			err,
		).WithContext("search", search)
	}
	
	return indicators, nil
}

// GetCountryCode finds ISO2 code by country name (case insensitive)
// Example: GetCountryCode("turkey") returns "TR", nil
// Example: GetCountryCode("United States") returns "US", nil
func GetCountryCode(countryName string) (string, error) {
	if countryName == "" {
		return "", NewPPPError(
			ErrCodeInvalidInput,
			"country name cannot be empty",
			nil,
		)
	}
	
	// Get all countries
	countries, err := ListCountries()
	if err != nil {
		return "", err
	}
	
	// Convert search term to lowercase for case-insensitive search
	searchName := strings.ToLower(strings.TrimSpace(countryName))
	
	// First try exact match
	for _, country := range countries {
		if strings.ToLower(country.Name) == searchName {
			return country.ISO2Code, nil
		}
	}
	
	// Then try contains match
	var matches []Country
	for _, country := range countries {
		if strings.Contains(strings.ToLower(country.Name), searchName) {
			matches = append(matches, country)
		}
	}
	
	// If exactly one match, return it
	if len(matches) == 1 {
		return matches[0].ISO2Code, nil
	}
	
	// If multiple matches, return error with suggestions
	if len(matches) > 1 {
		var suggestions []string
		for _, match := range matches {
			suggestions = append(suggestions, fmt.Sprintf("%s (%s)", match.Name, match.ISO2Code))
		}
		return "", NewPPPError(
			ErrCodeInvalidInput,
			fmt.Sprintf("multiple countries found for '%s'", countryName),
			nil,
		).WithContext("suggestions", suggestions)
	}
	
	// No matches found
	return "", NewPPPError(
		ErrCodeNoData,
		fmt.Sprintf("no country found for '%s'", countryName),
		nil,
	)
}

// BatchRecommend calculates recommended prices for multiple countries
// Returns (countryPrices, error)
func BatchRecommend(price float64, fromCurrency string, toCountries []string) (map[string]float64, error) {
	// Validate inputs
	if err := ValidateAmount(price); err != nil {
		return nil, err
	}
	
	if err := ValidateCurrencyCode(fromCurrency); err != nil {
		return nil, err
	}
	
	if len(toCountries) == 0 {
		return nil, NewPPPError(
			ErrCodeInvalidInput,
			"no countries provided",
			nil,
		)
	}
	
	// Validate all country codes
	for _, country := range toCountries {
		if err := ValidateCountryCode(country); err != nil {
			return nil, err
		}
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	results := make(map[string]float64)
	var lastError error
	
	for _, country := range toCountries {
		rec, err := defaultClient.Recommend(ctx, price, fromCurrency, country)
		if err != nil {
			// Store error but continue with other countries
			lastError = err
			results[country] = 0 // Mark as failed
		} else {
			results[country] = rec.RecommendedPrice
		}
	}
	
	// If all countries failed, return error
	allFailed := true
	for _, price := range results {
		if price > 0 {
			allFailed = false
			break
		}
	}
	
	if allFailed && lastError != nil {
		return nil, NewPPPError(
			ErrCodeAPIError,
			"failed to get recommendations for all countries",
			lastError,
		)
	}
	
	return results, nil
}

// GetPricingStrategy returns a pricing recommendation based on country
// Returns (recommendation, error)
func GetPricingStrategy(countryCode string, basePrice float64, baseCurrency string) (*PriceRecommendation, error) {
	return QuickRecommend(basePrice, baseCurrency, countryCode)
}

// DisableCache disables caching on the default client
func DisableCache() {
	defaultClient = NewClient(WithoutCache())
}

// EnableCache enables caching on the default client with specified duration
func EnableCache(duration time.Duration) {
	defaultClient = NewClient(WithCache(duration))
}