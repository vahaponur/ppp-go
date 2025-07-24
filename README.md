# ppp-go

A Go client library for working with Purchasing Power Parity (PPP) data. This library helps you implement location-based pricing by leveraging World Bank PPP data and real-time exchange rates.

## What is PPP?

Purchasing Power Parity (PPP) is an economic theory that compares different countries' currencies through a "basket of goods" approach. In simple terms, it tells you how much money you need in different countries to buy the same things.

For example, if a burger costs $5 in the US and ₺50 in Turkey, the PPP exchange rate would be 10 TRY/USD, even if the market exchange rate is different.

## Installation

```bash
go get github.com/vahaponur/ppp-go
```

## Quick Start

### First: Get Country Codes

You need 2-letter country codes (ISO 3166-1 alpha-2) to use this library.

#### Easy way - Search by country name:
```go
package main

import (
    "fmt"
    "log"
    "github.com/vahaponur/ppp-go"
)

func main() {
    // Find country code by name (case insensitive)
    code, err := ppp.GetCountryCode("turkey")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Country code: %s\n", code) // Output: Country code: TR
    
    // Works with partial names too
    code, _ = ppp.GetCountryCode("united states")
    fmt.Printf("US code: %s\n", code) // Output: US code: US
    
    code, _ = ppp.GetCountryCode("GERMANY")  // Case insensitive
    fmt.Printf("Germany code: %s\n", code) // Output: Germany code: DE
}
```

#### List all available countries:
```go
// Get all available countries
countries, err := ppp.ListCountries()
if err != nil {
    panic(err)
}

// Print first 10 countries
for i, country := range countries {
    if i >= 10 {
        break
    }
    fmt.Printf("%s: %s (Capital: %s)\n", 
        country.ISO2Code, country.Name, country.CapitalCity)
}

// Output:
// AW: Aruba (Capital: Oranjestad)
// AF: Afghanistan (Capital: Kabul)
// AO: Angola (Capital: Luanda)
// AI: Anguilla (Capital: The Valley)
// AL: Albania (Capital: Tirana)
// AD: Andorra (Capital: Andorra la Vella)
// AE: United Arab Emirates (Capital: Abu Dhabi)
// AR: Argentina (Capital: Buenos Aires)
// AM: Armenia (Capital: Yerevan)
// AS: American Samoa (Capital: Pago Pago)
```

### Simplest Usage - Get PPP-adjusted price

```go
// Get recommended price for Turkey
price, err := ppp.RecommendPrice(100, "USD", "TR")
if err != nil {
    panic(err)
}

fmt.Printf("Recommended price for Turkey: %.2f TRY\n", price)
// Output: Recommended price for Turkey: 1155.00 TRY
```

### Simple Functions

```go
// Get PPP factor for a country
factor, err := ppp.GetFactor("TR")
fmt.Printf("PPP Factor: %.2f\n", factor) // 11.55

// Get exchange rate
rate, err := ppp.GetRate("USD", "TRY")
fmt.Printf("Exchange Rate: %.2f\n", rate) // 40.47

// Convert price using PPP
localPrice, err := ppp.ConvertPrice(100, factor)
fmt.Printf("Local Price: %.2f\n", localPrice) // 1155.00
```

### Get Detailed Recommendation

```go
// Get detailed pricing recommendation
rec, err := ppp.QuickRecommend(99.99, "USD", "TR")
if err != nil {
    panic(err)
}

fmt.Printf("Original: $%.2f\n", rec.OriginalPrice)
fmt.Printf("Recommended: %.2f %s\n", rec.RecommendedPrice, rec.TargetCurrency)
fmt.Printf("Market Rate Price: %.2f\n", rec.OriginalPrice * rec.ExchangeRate)
fmt.Printf("PPP Discount: %.0f%%\n", rec.DiscountPercentage)

// Output:
// Original: $99.99
// Recommended: 1154.88 TRY
// Market Rate Price: 4046.60
// PPP Discount: 71%
```

## Advanced Usage

### Using the Client

```go
import (
    "context"
    "github.com/vahaponur/ppp-go"
)

// Create a client with custom options
client := ppp.NewClient(
    ppp.WithCache(24 * time.Hour),      // Enable caching
    ppp.WithTimeout(10 * time.Second),   // Set timeout
)

// Get PPP data
ctx := context.Background()
pppData, err := client.GetPPP(ctx, "TR")
if err != nil {
    panic(err)
}

fmt.Printf("Country: %s\n", pppData.CountryName)
fmt.Printf("PPP Factor: %.2f (Year: %d)\n", pppData.Factor, pppData.Year)
```

### Batch Pricing for Multiple Countries

```go
// Calculate prices for multiple countries at once
countries := []string{"TR", "BR", "IN", "MX", "ZA"}
prices, err := ppp.BatchRecommend(100, "USD", countries)
if err != nil {
    panic(err)
}

for country, price := range prices {
    fmt.Printf("%s: %.2f\n", country, price)
}

// Output:
// TR: 1155.00
// BR: 449.60
// IN: 236.80
// MX: 544.00
// ZA: 788.50
```

### SaaS Pricing Strategy

```go
// Create recommendation engine
engine := ppp.NewRecommendationEngine(ppp.NewClient())

// Get SaaS pricing recommendation
saas, err := engine.RecommendSaaS(ctx, 29.99, "USD", "TR")
if err != nil {
    panic(err)
}

fmt.Printf("Monthly: %.2f %s\n", saas.Monthly, saas.Currency)
fmt.Printf("Annual: %.2f %s (Save %.2f)\n", 
    saas.Annual, saas.Currency, saas.AnnualSavings)

// Output:
// Monthly: 346.38 TRY
// Annual: 3452.52 TRY (Save 703.04)
```

### Historical PPP Data

```go
// Get historical PPP data
history, err := client.GetHistoricalPPP(ctx, "TR", 2020, 2024)
if err != nil {
    panic(err)
}

for _, data := range history {
    fmt.Printf("%d: %.2f\n", data.Year, data.Factor)
}

// Output:
// 2024: 11.55
// 2023: 7.35
// 2022: 4.52
// 2021: 2.75
// 2020: 2.11
```

### Compare Countries

```go
// Compare PPP across multiple countries
countries := []string{"TR", "BR", "IN", "MX", "ZA", "PL"}
comparisons, err := client.ComparePPP(ctx, countries)
if err != nil {
    panic(err)
}

for _, comp := range comparisons {
    fmt.Printf("Rank %d: %s - PPP: %.2f (%.0f%% of US purchasing power)\n",
        comp.Rank, comp.CountryName, comp.Factor, comp.PercentOfUS)
}
```

### Price Formatting

```go
// Format prices according to currency conventions
price := 1234.56

fmt.Println(ppp.FormatPrice(price, "USD"))  // $1234.56
fmt.Println(ppp.FormatPrice(price, "EUR"))  // 1234.56 €
fmt.Println(ppp.FormatPrice(price, "TRY"))  // ₺1234.56
fmt.Println(ppp.FormatPrice(price, "JPY"))  // ¥1235
```

## Features

### 🌍 Multiple Data Sources
- **World Bank API**: Official PPP data (PA.NUS.PPP indicator)
- **Currency Exchange API**: Real-time exchange rates
- **Historical Data**: PPP trends over time

### 💰 Pricing Tools
- **Simple Price Conversion**: One-line price recommendations
- **Detailed Analysis**: PPP factors, exchange rates, and discount percentages
- **SaaS Pricing**: Monthly/annual pricing with PPP adjustments
- **Batch Operations**: Calculate prices for multiple countries

### 🚀 Performance
- **Built-in Caching**: Reduce API calls and improve response times
- **Configurable Timeouts**: Control API request timeouts
- **Error Handling**: Comprehensive error types with context

### 🎯 Easy to Use
- **Zero-config**: Works out of the box with sensible defaults
- **Package-level Functions**: Simple functions for common tasks
- **Flexible Client**: Advanced features when you need them

## API Sources

1. **World Bank API**
   - Base URL: `https://api.worldbank.org/v2`
   - No authentication required
   - Provides official PPP data

2. **Currency Exchange API**
   - Base URL: `https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api@latest/v1`
   - No authentication required
   - Provides real-time exchange rates

## Configuration

### Disable Caching
```go
// Disable cache globally
ppp.DisableCache()

// Or create client without cache
client := ppp.NewClient(ppp.WithoutCache())
```

### Custom API URLs
```go
client := ppp.NewClient(
    ppp.WithWorldBankURL("https://custom.worldbank.api"),
    ppp.WithCurrencyURL("https://custom.currency.api"),
)
```

### Cache Management
```go
// Enable cache with custom duration
ppp.EnableCache(1 * time.Hour)

// Export cache to file
client.ExportCache("cache.json")

// Import cache from file
client.ImportCache("cache.json")

// Clear cache
client.ClearCache()
```

## Error Handling

The library provides detailed error information:

```go
price, err := ppp.RecommendPrice(100, "USD", "XX")
if err != nil {
    var pppErr *ppp.PPPError
    if errors.As(err, &pppErr) {
        fmt.Printf("Error Code: %s\n", pppErr.Code)
        fmt.Printf("Message: %s\n", pppErr.Message)
        fmt.Printf("Context: %v\n", pppErr.Context)
    }
}
```

Error codes:
- `NO_DATA`: No data available
- `API_ERROR`: API request failed
- `INVALID_INPUT`: Invalid input parameters
- `TIMEOUT`: Request timeout
- `NETWORK_ERROR`: Network issue

## Common Use Cases

### E-commerce Localized Pricing
```go
func GetLocalizedPrice(usdPrice float64, countryCode string) string {
    price, err := ppp.RecommendPrice(usdPrice, "USD", countryCode)
    if err != nil {
        // Fallback to USD price
        return fmt.Sprintf("$%.2f", usdPrice)
    }
    
    currency := ppp.GetCurrencyForCountry(countryCode)
    return ppp.FormatPrice(price, currency)
}
```

### SaaS Tier Pricing
```go
tiers := map[string]float64{
    "Basic": 9.99,
    "Pro": 29.99,
    "Enterprise": 99.99,
}

countryPrices := make(map[string]map[string]float64)
for tier, usdPrice := range tiers {
    for _, country := range []string{"TR", "BR", "IN"} {
        price, _ := ppp.RecommendPrice(usdPrice, "USD", country)
        if countryPrices[country] == nil {
            countryPrices[country] = make(map[string]float64)
        }
        countryPrices[country][tier] = price
    }
}
```

### Dynamic Pricing API Endpoint
```go
func PricingHandler(w http.ResponseWriter, r *http.Request) {
    countryCode := r.URL.Query().Get("country")
    productPrice := 49.99
    
    rec, err := ppp.QuickRecommend(productPrice, "USD", countryCode)
    if err != nil {
        http.Error(w, "Pricing unavailable", http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "original_price": rec.OriginalPrice,
        "local_price": rec.RecommendedPrice,
        "currency": rec.TargetCurrency,
        "discount": rec.DiscountPercentage,
    })
}
```

## Finding Country Codes

### Search for a specific country:
```go
countries, _ := ppp.ListCountries()
for _, country := range countries {
    // Search by name (case insensitive)
    if strings.Contains(strings.ToLower(country.Name), "turkey") {
        fmt.Printf("Found: %s = %s\n", country.ISO2Code, country.Name)
        // Output: Found: TR = Turkiye
    }
}
```

### Get country details:
```go
client := ppp.NewClient()
countries, _ := client.GetCountries(context.Background())

// Find specific country
for _, c := range countries {
    if c.ISO2Code == "TR" {
        fmt.Printf("Country: %s\n", c.Name)
        fmt.Printf("Code: %s\n", c.ISO2Code)
        fmt.Printf("Capital: %s\n", c.CapitalCity)
        fmt.Printf("Region: %s\n", c.Region.Value)
        fmt.Printf("Income Level: %s\n", c.IncomeLevel.Value)
        break
    }
}

// Output:
// Country: Turkiye
// Code: TR
// Capital: Ankara
// Region: Europe & Central Asia
// Income Level: Upper middle income
```

## Testing

```bash
# Run all tests
go test -v ./...

# Run integration tests (requires internet)
go test -v -run TestRealAPI

# Run specific test
go test -v -run TestRealAPI/GetFactor_Turkey

# Run benchmarks
go test -bench=. -run=XXX
```

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

For issues and feature requests, please use the [GitHub issue tracker](https://github.com/vahaponur/ppp-go/issues).