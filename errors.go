package ppp

import (
	"errors"
	"fmt"
	"time"
)

// Error codes for PPP operations
const (
	ErrCodeNoData       = "NO_DATA"
	ErrCodeAPIError     = "API_ERROR"
	ErrCodeInvalidInput = "INVALID_INPUT"
	ErrCodeTimeout      = "TIMEOUT"
	ErrCodeRateLimit    = "RATE_LIMIT"
	ErrCodeCacheError   = "CACHE_ERROR"
	ErrCodeNetworkError = "NETWORK_ERROR"
)

// Common errors
var (
	ErrNoData            = errors.New("no data available")
	ErrInvalidCountry    = errors.New("invalid country code")
	ErrInvalidCurrency   = errors.New("invalid currency code")
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrAPIUnavailable    = errors.New("API service unavailable")
	ErrCacheDisabled     = errors.New("cache is disabled")
	ErrInvalidDateRange  = errors.New("invalid date range")
)

// PPPError represents a detailed error with code and context
type PPPError struct {
	Code    string
	Message string
	Err     error
	Context map[string]interface{}
}

// Error implements the error interface
func (e *PPPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *PPPError) Unwrap() error {
	return e.Err
}

// NewPPPError creates a new PPP error
func NewPPPError(code, message string, err error) *PPPError {
	return &PPPError{
		Code:    code,
		Message: message,
		Err:     err,
		Context: make(map[string]interface{}),
	}
}

// WithContext adds context to the error
func (e *PPPError) WithContext(key string, value interface{}) *PPPError {
	e.Context[key] = value
	return e
}

// IsNoDataError checks if error is a no data error
func IsNoDataError(err error) bool {
	var pppErr *PPPError
	if errors.As(err, &pppErr) {
		return pppErr.Code == ErrCodeNoData
	}
	return errors.Is(err, ErrNoData)
}

// IsAPIError checks if error is an API error
func IsAPIError(err error) bool {
	var pppErr *PPPError
	if errors.As(err, &pppErr) {
		return pppErr.Code == ErrCodeAPIError
	}
	return false
}

// IsNetworkError checks if error is a network error
func IsNetworkError(err error) bool {
	var pppErr *PPPError
	if errors.As(err, &pppErr) {
		return pppErr.Code == ErrCodeNetworkError
	}
	return false
}

// ValidateCountryCode validates country code format
func ValidateCountryCode(code string) error {
	if len(code) != 2 {
		return NewPPPError(
			ErrCodeInvalidInput,
			"country code must be 2 characters",
			ErrInvalidCountry,
		).WithContext("country_code", code)
	}
	
	// Check if it's uppercase
	for _, char := range code {
		if char < 'A' || char > 'Z' {
			return NewPPPError(
				ErrCodeInvalidInput,
				"country code must be uppercase letters",
				ErrInvalidCountry,
			).WithContext("country_code", code)
		}
	}
	
	return nil
}

// ValidateCurrencyCode validates currency code format
func ValidateCurrencyCode(code string) error {
	if len(code) != 3 {
		return NewPPPError(
			ErrCodeInvalidInput,
			"currency code must be 3 characters",
			ErrInvalidCurrency,
		).WithContext("currency_code", code)
	}
	
	// Check if it's uppercase
	for _, char := range code {
		if char < 'A' || char > 'Z' {
			return NewPPPError(
				ErrCodeInvalidInput,
				"currency code must be uppercase letters",
				ErrInvalidCurrency,
			).WithContext("currency_code", code)
		}
	}
	
	return nil
}

// ValidateAmount validates monetary amount
func ValidateAmount(amount float64) error {
	if amount < 0 {
		return NewPPPError(
			ErrCodeInvalidInput,
			"amount cannot be negative",
			ErrInvalidAmount,
		).WithContext("amount", amount)
	}
	
	if amount == 0 {
		return NewPPPError(
			ErrCodeInvalidInput,
			"amount cannot be zero",
			ErrInvalidAmount,
		).WithContext("amount", amount)
	}
	
	// Check for very large amounts that might cause overflow
	if amount > 1e15 {
		return NewPPPError(
			ErrCodeInvalidInput,
			"amount is too large",
			ErrInvalidAmount,
		).WithContext("amount", amount)
	}
	
	return nil
}

// ValidateDateRange validates year range
func ValidateDateRange(startYear, endYear int) error {
	currentYear := getCurrentYear()
	
	if startYear < 1960 {
		return NewPPPError(
			ErrCodeInvalidInput,
			"start year cannot be before 1960 (World Bank data limitation)",
			ErrInvalidDateRange,
		).WithContext("start_year", startYear)
	}
	
	if endYear > currentYear+1 {
		return NewPPPError(
			ErrCodeInvalidInput,
			"end year cannot be in the future",
			ErrInvalidDateRange,
		).WithContext("end_year", endYear).WithContext("current_year", currentYear)
	}
	
	if startYear > endYear {
		return NewPPPError(
			ErrCodeInvalidInput,
			"start year must be before or equal to end year",
			ErrInvalidDateRange,
		).WithContext("start_year", startYear).WithContext("end_year", endYear)
	}
	
	return nil
}

// getCurrentYear is a helper to get current year
func getCurrentYear() int {
	return time.Now().Year()
}