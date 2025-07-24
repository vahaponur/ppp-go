package ppp

import (
	"context"
	"fmt"
	"math"
)

// PricingTier represents different pricing tiers based on PPP
type PricingTier struct {
	Name               string
	MinPPPFactor       float64
	MaxPPPFactor       float64
	DiscountPercentage float64
}

// StandardPricingTiers provides common pricing tiers
var StandardPricingTiers = []PricingTier{
	{"Premium", 0, 0.3, 70},       // Very low purchasing power
	{"Standard", 0.3, 0.6, 50},    // Low purchasing power
	{"Regular", 0.6, 0.9, 25},     // Medium purchasing power
	{"Full Price", 0.9, 999, 0},   // High purchasing power
}

// RecommendationEngine provides advanced price recommendation logic
type RecommendationEngine struct {
	client       *Client
	pricingTiers []PricingTier
}

// NewRecommendationEngine creates a new recommendation engine
func NewRecommendationEngine(client *Client) *RecommendationEngine {
	return &RecommendationEngine{
		client:       client,
		pricingTiers: StandardPricingTiers,
	}
}

// SetPricingTiers sets custom pricing tiers
func (r *RecommendationEngine) SetPricingTiers(tiers []PricingTier) {
	r.pricingTiers = tiers
}

// RecommendWithStrategy provides strategic price recommendation
func (r *RecommendationEngine) RecommendWithStrategy(ctx context.Context, price float64, fromCurrency, toCountry string) (*PriceRecommendation, error) {
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
	
	// Get base recommendation
	rec, err := r.client.Recommend(ctx, price, fromCurrency, toCountry)
	if err != nil {
		return nil, err
	}
	
	// Apply tiered pricing strategy
	tier := r.getTierForPPP(rec.PPPFactor)
	if tier != nil {
		// Apply tier discount
		rec.RecommendedPrice = price * (1 - tier.DiscountPercentage/100)
		rec.DiscountPercentage = tier.DiscountPercentage
	}
	
	return rec, nil
}

// getTierForPPP finds the appropriate pricing tier
func (r *RecommendationEngine) getTierForPPP(pppFactor float64) *PricingTier {
	for _, tier := range r.pricingTiers {
		if pppFactor >= tier.MinPPPFactor && pppFactor < tier.MaxPPPFactor {
			return &tier
		}
	}
	return nil
}

// RecommendSaaS provides SaaS-specific pricing recommendation
func (r *RecommendationEngine) RecommendSaaS(ctx context.Context, monthlyPrice float64, fromCurrency, toCountry string) (*SaaSPricing, error) {
	// Get base recommendation
	rec, err := r.RecommendWithStrategy(ctx, monthlyPrice, fromCurrency, toCountry)
	if err != nil {
		return nil, err
	}
	
	// Calculate annual pricing with discount
	annualDiscount := 0.167 // ~2 months free
	annualPrice := rec.RecommendedPrice * 12 * (1 - annualDiscount)
	
	return &SaaSPricing{
		Monthly:            rec.RecommendedPrice,
		Annual:             annualPrice,
		AnnualSavings:      (rec.RecommendedPrice * 12) - annualPrice,
		Currency:           rec.TargetCurrency,
		DiscountPercentage: rec.DiscountPercentage,
		PPPAdjusted:        true,
	}, nil
}

// SaaSPricing represents SaaS-specific pricing
type SaaSPricing struct {
	Monthly            float64
	Annual             float64
	AnnualSavings      float64
	Currency           string
	DiscountPercentage float64
	PPPAdjusted        bool
}

// RoundPrice rounds price to appropriate decimal places based on currency
func RoundPrice(price float64, currency string) float64 {
	// Currencies that typically don't use decimals
	noDecimalCurrencies := map[string]bool{
		"JPY": true,
		"KRW": true,
		"IDR": true,
		"VND": true,
		"CLP": true,
		"PYG": true,
		"RWF": true,
		"XAF": true,
		"XOF": true,
		"XPF": true,
	}
	
	if noDecimalCurrencies[currency] {
		return math.Round(price)
	}
	
	// Round to 2 decimal places for most currencies
	return math.Round(price*100) / 100
}

// FormatPrice formats price according to currency conventions
func FormatPrice(price float64, currency string) string {
	rounded := RoundPrice(price, currency)
	
	currencySymbols := map[string]string{
		"USD": "$",
		"EUR": "€",
		"GBP": "£",
		"JPY": "¥",
		"CNY": "¥",
		"INR": "₹",
		"TRY": "₺",
		"RUB": "₽",
		"KRW": "₩",
		"BRL": "R$",
		"MXN": "$",
		"CAD": "C$",
		"AUD": "A$",
		"CHF": "CHF",
		"SEK": "kr",
		"NOK": "kr",
		"DKK": "kr",
		"PLN": "zł",
		"CZK": "Kč",
		"HUF": "Ft",
		"ILS": "₪",
		"AED": "د.إ",
		"SAR": "ر.س",
		"ZAR": "R",
		"NGN": "₦",
		"EGP": "E£",
		"PKR": "₨",
		"BDT": "৳",
		"VND": "₫",
		"THB": "฿",
		"MYR": "RM",
		"SGD": "S$",
		"PHP": "₱",
		"IDR": "Rp",
		"NZD": "NZ$",
		"ARS": "$",
		"CLP": "$",
		"COP": "$",
		"PEN": "S/",
	}
	
	symbol, ok := currencySymbols[currency]
	if !ok {
		symbol = currency + " "
	}
	
	// Format based on currency conventions
	switch currency {
	case "EUR", "RUB", "PLN", "CZK", "HUF":
		// Symbol after amount for these currencies
		return fmt.Sprintf("%.2f %s", rounded, symbol)
	case "JPY", "KRW", "IDR", "VND", "CLP":
		// No decimals for these currencies
		return fmt.Sprintf("%s%.0f", symbol, rounded)
	default:
		// Symbol before amount (most common)
		return fmt.Sprintf("%s%.2f", symbol, rounded)
	}
}

// CalculateMarketBasket calculates PPP-adjusted prices for multiple items
func CalculateMarketBasket(ctx context.Context, client *Client, items map[string]float64, fromCurrency, toCountry string) (map[string]float64, error) {
	// Validate inputs
	if err := ValidateCurrencyCode(fromCurrency); err != nil {
		return nil, err
	}
	
	if err := ValidateCountryCode(toCountry); err != nil {
		return nil, err
	}
	
	if len(items) == 0 {
		return nil, NewPPPError(
			ErrCodeInvalidInput,
			"no items provided",
			nil,
		)
	}
	
	// Get PPP data once
	ppp, err := client.GetPPP(ctx, toCountry)
	if err != nil {
		return nil, err
	}
	
	// Get currency for country
	toCurrency := client.getCurrencyForCountry(toCountry)
	
	// Get exchange rate once
	rate, err := client.GetExchangeRate(ctx, fromCurrency, toCurrency)
	if err != nil {
		return nil, err
	}
	
	// Calculate adjusted prices for all items
	results := make(map[string]float64)
	for item, price := range items {
		if err := ValidateAmount(price); err != nil {
			return nil, fmt.Errorf("invalid price for item %s: %w", item, err)
		}
		
		// Apply PPP adjustment
		adjustedPrice := price * (ppp.Factor / rate.Rate)
		results[item] = RoundPrice(adjustedPrice, toCurrency)
	}
	
	return results, nil
}