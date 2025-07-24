package ppp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/patrickmn/go-cache"
)

// Cache provides a simple caching layer for PPP data
type Cache struct {
	memory *cache.Cache
}

// NewCache creates a new cache instance
func NewCache(defaultExpiration, cleanupInterval time.Duration) *Cache {
	return &Cache{
		memory: cache.New(defaultExpiration, cleanupInterval),
	}
}

// CacheKey generates a cache key for PPP data
func CacheKeyPPP(countryCode string) string {
	return fmt.Sprintf("ppp:%s", countryCode)
}

// CacheKeyExchangeRate generates a cache key for exchange rate
func CacheKeyExchangeRate(from, to string) string {
	return fmt.Sprintf("rate:%s:%s", from, to)
}

// CacheKeyCountries generates a cache key for countries list
func CacheKeyCountries() string {
	return "countries:all"
}

// CacheKeyIndicators generates a cache key for indicators search
func CacheKeyIndicators(search string) string {
	return fmt.Sprintf("indicators:search:%s", search)
}

// GetPPP retrieves PPP data from cache
func (c *Cache) GetPPP(countryCode string) (*PPPData, bool) {
	key := CacheKeyPPP(countryCode)
	if data, found := c.memory.Get(key); found {
		if ppp, ok := data.(*PPPData); ok {
			return ppp, true
		}
	}
	return nil, false
}

// SetPPP stores PPP data in cache
func (c *Cache) SetPPP(countryCode string, data *PPPData, expiration time.Duration) {
	key := CacheKeyPPP(countryCode)
	c.memory.Set(key, data, expiration)
}

// GetExchangeRate retrieves exchange rate from cache
func (c *Cache) GetExchangeRate(from, to string) (*ExchangeRate, bool) {
	key := CacheKeyExchangeRate(from, to)
	if data, found := c.memory.Get(key); found {
		if rate, ok := data.(*ExchangeRate); ok {
			return rate, true
		}
	}
	return nil, false
}

// SetExchangeRate stores exchange rate in cache
func (c *Cache) SetExchangeRate(from, to string, rate *ExchangeRate, expiration time.Duration) {
	key := CacheKeyExchangeRate(from, to)
	c.memory.Set(key, rate, expiration)
}

// GetCountries retrieves countries list from cache
func (c *Cache) GetCountries() ([]Country, bool) {
	key := CacheKeyCountries()
	if data, found := c.memory.Get(key); found {
		if countries, ok := data.([]Country); ok {
			return countries, true
		}
	}
	return nil, false
}

// SetCountries stores countries list in cache
func (c *Cache) SetCountries(countries []Country, expiration time.Duration) {
	key := CacheKeyCountries()
	c.memory.Set(key, countries, expiration)
}

// GetIndicators retrieves indicators from cache
func (c *Cache) GetIndicators(search string) ([]Indicator, bool) {
	key := CacheKeyIndicators(search)
	if data, found := c.memory.Get(key); found {
		if indicators, ok := data.([]Indicator); ok {
			return indicators, true
		}
	}
	return nil, false
}

// SetIndicators stores indicators in cache
func (c *Cache) SetIndicators(search string, indicators []Indicator, expiration time.Duration) {
	key := CacheKeyIndicators(search)
	c.memory.Set(key, indicators, expiration)
}

// Clear removes all items from cache
func (c *Cache) Clear() {
	c.memory.Flush()
}

// ExportToFile exports cache data to a JSON file
func (c *Cache) ExportToFile(filename string) error {
	items := c.memory.Items()
	
	// Convert to a serializable format
	exportData := make(map[string]interface{})
	for k, v := range items {
		exportData[k] = v.Object
	}
	
	data, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}
	
	return nil
}

// ImportFromFile imports cache data from a JSON file
func (c *Cache) ImportFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read cache file: %w", err)
	}
	
	var importData map[string]json.RawMessage
	if err := json.Unmarshal(data, &importData); err != nil {
		return fmt.Errorf("failed to unmarshal cache data: %w", err)
	}
	
	// Import each item based on its key prefix
	for key, rawData := range importData {
		switch {
		case len(key) > 4 && key[:4] == "ppp:":
			var ppp PPPData
			if err := json.Unmarshal(rawData, &ppp); err == nil {
				c.memory.Set(key, &ppp, cache.DefaultExpiration)
			}
		case len(key) > 5 && key[:5] == "rate:":
			var rate ExchangeRate
			if err := json.Unmarshal(rawData, &rate); err == nil {
				c.memory.Set(key, &rate, cache.DefaultExpiration)
			}
		case key == "countries:all":
			var countries []Country
			if err := json.Unmarshal(rawData, &countries); err == nil {
				c.memory.Set(key, countries, cache.DefaultExpiration)
			}
		case len(key) > 17 && key[:17] == "indicators:search":
			var indicators []Indicator
			if err := json.Unmarshal(rawData, &indicators); err == nil {
				c.memory.Set(key, indicators, cache.DefaultExpiration)
			}
		}
	}
	
	return nil
}