package ppp

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRealAPI(t *testing.T) {
	// Skip if short test
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Test 1: Basit PPP factor çekme
	t.Run("GetFactor_Turkey", func(t *testing.T) {
		factor, err := GetFactor("TR")
		if err != nil {
			t.Fatalf("GetFactor failed: %v", err)
		}

		// PPP factor 0'dan büyük olmalı
		if factor <= 0 {
			t.Errorf("Invalid PPP factor: %v", factor)
		}

		t.Logf("Turkey PPP Factor: %.2f", factor)
	})

	// Test 2: Exchange rate çekme
	t.Run("GetRate_USD_TRY", func(t *testing.T) {
		rate, err := GetRate("USD", "TRY")
		if err != nil {
			t.Fatalf("GetRate failed: %v", err)
		}

		// USD/TRY 20'den büyük olmalı (2024 için)
		if rate < 20 {
			t.Errorf("Suspicious USD/TRY rate: %v", rate)
		}

		t.Logf("USD/TRY Rate: %.2f", rate)
	})

	// Test 3: Fiyat önerisi
	t.Run("RecommendPrice_100USD_to_Turkey", func(t *testing.T) {
		price, err := RecommendPrice(100, "USD", "TR")
		if err != nil {
			t.Fatalf("RecommendPrice failed: %v", err)
		}

		// Önerilen fiyat 0'dan büyük olmalı
		if price <= 0 {
			t.Errorf("Invalid recommended price: %v", price)
		}

		t.Logf("Recommended price for $100 in Turkey: %.2f", price)
	})

	// Test 4: QuickRecommend ile detaylı bilgi
	t.Run("QuickRecommend_SaaS_Pricing", func(t *testing.T) {
		rec, err := QuickRecommend(50, "USD", "TR")
		if err != nil {
			t.Fatalf("QuickRecommend failed: %v", err)
		}

		t.Logf("Original: $%.2f", rec.OriginalPrice)
		t.Logf("Recommended: %.2f %s", rec.RecommendedPrice, rec.TargetCurrency)
		t.Logf("PPP Factor: %.2f", rec.PPPFactor)
		t.Logf("Exchange Rate: %.2f", rec.ExchangeRate)
		t.Logf("Discount: %.0f%%", rec.DiscountPercentage)

		// Mantıklı değerler kontrolü
		if rec.PPPFactor <= 0 || rec.PPPFactor > 50 {
			t.Errorf("Suspicious PPP factor: %v", rec.PPPFactor)
		}

		if rec.ExchangeRate < 20 || rec.ExchangeRate > 100 {
			t.Errorf("Suspicious exchange rate: %v", rec.ExchangeRate)
		}
	})

	// Test 5: Batch recommend
	t.Run("BatchRecommend_Multiple_Countries", func(t *testing.T) {
		countries := []string{"TR", "BR", "IN", "MX"}
		prices, err := BatchRecommend(100, "USD", countries)
		if err != nil {
			t.Fatalf("BatchRecommend failed: %v", err)
		}

		for country, price := range prices {
			if price <= 0 {
				t.Errorf("Invalid price for %s: %v", country, price)
			}
			t.Logf("%s: %.2f", country, price)
		}
	})

	// Test 6: Client ile detaylı test
	t.Run("Client_Direct_API_Calls", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()

		// PPP data çek
		ppp, err := client.GetPPP(ctx, "TR")
		if err != nil {
			t.Fatalf("Client.GetPPP failed: %v", err)
		}

		t.Logf("Turkey PPP Data: Country=%s, Year=%d, Factor=%.2f",
			ppp.CountryName, ppp.Year, ppp.Factor)

		// Exchange rate çek
		rate, err := client.GetExchangeRate(ctx, "USD", "TRY")
		if err != nil {
			t.Fatalf("Client.GetExchangeRate failed: %v", err)
		}

		t.Logf("Exchange Rate: %s to %s = %.2f (Updated: %s)",
			rate.From, rate.To, rate.Rate, rate.LastUpdated.Format("2006-01-02"))

		// Recommendation
		rec, err := client.Recommend(ctx, 99.99, "USD", "TR")
		if err != nil {
			t.Fatalf("Client.Recommend failed: %v", err)
		}

		t.Logf("Price Recommendation: $%.2f -> %.2f %s (%.0f%% discount)",
			rec.OriginalPrice, rec.RecommendedPrice, rec.TargetCurrency, rec.DiscountPercentage)
	})

	// Test 7: Historical data
	t.Run("Historical_PPP_Data", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()

		currentYear := time.Now().Year()
		history, err := client.GetHistoricalPPP(ctx, "TR", currentYear-5, currentYear)
		if err != nil {
			t.Fatalf("GetHistoricalPPP failed: %v", err)
		}

		if len(history) == 0 {
			t.Error("No historical data returned")
		}

		for _, h := range history {
			t.Logf("Year %d: %.2f", h.Year, h.Factor)
		}
	})

	// Test 8: Country list
	t.Run("List_Countries", func(t *testing.T) {
		countries, err := ListCountries()
		if err != nil {
			t.Fatalf("ListCountries failed: %v", err)
		}
		if len(countries) == 0 {
			t.Error("No countries returned")
		}

		// Türkiye'yi bul
		foundTurkey := false
		for _, c := range countries {
			if c.ISO2Code == "TR" {
				foundTurkey = true
				t.Logf("Found Turkey: %s, Capital: %s, Region: %s",
					c.Name, c.CapitalCity, c.Region.Value)
				break
			}
		}

		if !foundTurkey {
			t.Error("Turkey not found in country list")
		}

		t.Logf("Total countries: %d", len(countries))
	})

	// Test 9: Search indicators
	t.Run("Search_Indicators", func(t *testing.T) {
		indicators, err := FindIndicators("gdp per capita")
		if err != nil {
			t.Fatalf("FindIndicators failed: %v", err)
		}

		if len(indicators) == 0 {
			t.Error("No indicators found")
		}

		// İlk 3 indicator'ı logla
		for i, ind := range indicators {
			if i >= 3 {
				break
			}
			t.Logf("Indicator: %s - %s", ind.ID, ind.Name)
		}
	})

	// Test 10: Format price
	t.Run("Format_Prices", func(t *testing.T) {
		testCases := []struct {
			price    float64
			currency string
		}{
			{99.99, "USD"},
			{99.99, "TRY"},
			{1000, "JPY"},
			{50000, "KRW"},
		}

		for _, tc := range testCases {
			formatted := FormatPrice(tc.price, tc.currency)
			t.Logf("%v %s -> %s", tc.price, tc.currency, formatted)
		}
	})

	// Test 11: PPP Trend Analysis
	t.Run("PPP_Trend_Analysis", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()
		
		currentYear := time.Now().Year()
		analysis, err := client.AnalyzePPPTrend(ctx, "TR", currentYear-10, currentYear)
		if err != nil {
			t.Fatalf("AnalyzePPPTrend failed: %v", err)
		}
		
		t.Logf("PPP Trend Analysis for Turkey (%d-%d):", analysis.StartYear, analysis.EndYear)
		t.Logf("  Average PPP: %.2f", analysis.Average)
		t.Logf("  Trend: %s", analysis.Trend)
		t.Logf("  Volatility: %.2f", analysis.Volatility)
		t.Logf("  Data Points: %d", analysis.DataPoints)
	})

	// Test 12: Compare PPP across countries
	t.Run("Compare_PPP_Countries", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()
		
		countries := []string{"TR", "BR", "IN", "MX", "ZA", "PL"}
		comparisons, err := client.ComparePPP(ctx, countries)
		if err != nil {
			t.Fatalf("ComparePPP failed: %v", err)
		}
		
		t.Log("PPP Comparison (sorted by purchasing power):")
		for _, comp := range comparisons {
			t.Logf("  Rank %d: %s (%s) - Factor: %.2f, %% of US: %.1f%%", 
				comp.Rank, comp.Country, comp.CountryName, comp.Factor, comp.PercentOfUS)
		}
	})

	// Test 13: Pricing Strategy with Tiers
	t.Run("Pricing_Strategy_SaaS", func(t *testing.T) {
		engine := NewRecommendationEngine(NewClient())
		ctx := context.Background()
		
		saasPrice, err := engine.RecommendSaaS(ctx, 29.99, "USD", "TR")
		if err != nil {
			t.Fatalf("RecommendSaaS failed: %v", err)
		}
		
		t.Logf("SaaS Pricing for Turkey:")
		t.Logf("  Monthly: %.2f %s", saasPrice.Monthly, saasPrice.Currency)
		t.Logf("  Annual: %.2f %s (Save %.2f)", saasPrice.Annual, saasPrice.Currency, saasPrice.AnnualSavings)
		t.Logf("  PPP Discount: %.0f%%", saasPrice.DiscountPercentage)
	})

	// Test 14: Market Basket calculation
	t.Run("Market_Basket_Pricing", func(t *testing.T) {
		client := NewClient()
		ctx := context.Background()
		
		items := map[string]float64{
			"Basic Plan":    9.99,
			"Pro Plan":      29.99,
			"Enterprise":    99.99,
		}
		
		basket, err := CalculateMarketBasket(ctx, client, items, "USD", "TR")
		if err != nil {
			t.Fatalf("CalculateMarketBasket failed: %v", err)
		}
		
		t.Log("Market Basket PPP Pricing for Turkey:")
		for item, price := range basket {
			t.Logf("  %s: %.2f TRY", item, price)
		}
	})

	// Test 15: Cache functionality
	t.Run("Cache_Operations", func(t *testing.T) {
		// Test with cache
		clientWithCache := NewClient(WithCache(5 * time.Minute))
		ctx := context.Background()
		
		// First call - should hit API
		start := time.Now()
		ppp1, err := clientWithCache.GetPPP(ctx, "BR")
		if err != nil {
			t.Fatalf("First GetPPP failed: %v", err)
		}
		duration1 := time.Since(start)
		
		// Second call - should hit cache
		start = time.Now()
		ppp2, err := clientWithCache.GetPPP(ctx, "BR")
		if err != nil {
			t.Fatalf("Second GetPPP failed: %v", err)
		}
		duration2 := time.Since(start)
		
		t.Logf("Cache test results:")
		t.Logf("  First call (API): %.2f ms", duration1.Seconds()*1000)
		t.Logf("  Second call (Cache): %.2f ms", duration2.Seconds()*1000)
		t.Logf("  Speed improvement: %.0fx faster", duration1.Seconds()/duration2.Seconds())
		
		if ppp1.Factor != ppp2.Factor {
			t.Error("Cache returned different value")
		}
	})

	// Test 16: Error handling
	t.Run("Error_Handling", func(t *testing.T) {
		// Test invalid country code
		_, err := GetFactor("XX")
		if err == nil {
			t.Error("Expected error for invalid country code")
		}
		
		// Test invalid currency
		_, err = GetRate("XXX", "YYY")
		if err == nil {
			t.Error("Expected error for invalid currency")
		}
		
		// Test negative amount
		_, err = RecommendPrice(-100, "USD", "TR")
		if err == nil {
			t.Error("Expected error for negative amount")
		}
		
		var pppErr *PPPError
		if errors.As(err, &pppErr) {
			t.Logf("Error handling works: Code=%s, Message=%s", pppErr.Code, pppErr.Message)
		}
	})
}

// Benchmark testleri
func BenchmarkGetFactor(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GetFactor("TR")
	}
}

func BenchmarkRecommendPrice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = RecommendPrice(100, "USD", "TR")
	}
}

func BenchmarkBatchRecommend(b *testing.B) {
	countries := []string{"TR", "BR", "IN", "MX", "ZA"}
	for i := 0; i < b.N; i++ {
		_, _ = BatchRecommend(100, "USD", countries)
	}
}
