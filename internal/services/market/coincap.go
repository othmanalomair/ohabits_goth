package market

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// CoinCapV3Response represents the response from CoinCap API v3
type CoinCapV3Response struct {
	Data      []CoinCapV3Asset `json:"data"`
	Timestamp int64            `json:"timestamp"`
}

// CoinCapV3Asset represents a cryptocurrency asset from CoinCap v3 API
type CoinCapV3Asset struct {
	ID                string  `json:"id"`                // slug like "bitcoin"
	Rank              string  `json:"rank"`              // returned as string
	Symbol            string  `json:"symbol"`            // "BTC"
	Name              string  `json:"name"`              // "Bitcoin"
	Supply            string  `json:"supply"`            // returned as string
	MaxSupply         *string `json:"maxSupply"`         // can be null, returned as string
	MarketCapUsd      string  `json:"marketCapUsd"`      // returned as string
	VolumeUsd24Hr     string  `json:"volumeUsd24Hr"`     // returned as string
	PriceUsd          string  `json:"priceUsd"`          // returned as string
	ChangePercent24Hr string  `json:"changePercent24Hr"` // returned as string
	Vwap24Hr          *string `json:"vwap24Hr"`          // can be null, returned as string
}

// MarketData represents processed market data for our database
type MarketData struct {
	Symbol        string
	CurrentPrice  float64
	ChangeAmount  float64
	ChangePercent float64
	Volume        int64
	MarketCap     int64
	LastUpdated   time.Time
}

// CoinCapClient handles communication with the CoinCap API
type CoinCapClient struct {
	client *http.Client
	apiKey string
	baseURL string
}

// NewCoinCapClient creates a new CoinCap API client
func NewCoinCapClient() *CoinCapClient {
	return &CoinCapClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey:  os.Getenv("COINCAP_API_KEY"),
		baseURL: "https://rest.coincap.io/v3",
	}
}

// GetAssets fetches cryptocurrency data for specified symbols
func (c *CoinCapClient) GetAssets(symbols []string) ([]MarketData, error) {
	// Map symbols to their corresponding CoinCap slugs
	symbolToSlug := map[string]string{
		"BTC":  "bitcoin",
		"ETH":  "ethereum", 
		"XRP":  "xrp",
		"DOGE": "dogecoin",
	}
	
	// Build comma-separated list of slugs
	var slugs []string
	for _, symbol := range symbols {
		if slug, exists := symbolToSlug[symbol]; exists {
			slugs = append(slugs, slug)
		}
	}
	
	if len(slugs) == 0 {
		return []MarketData{}, nil
	}
	
	// Create URL with ids parameter for v3 API
	url := fmt.Sprintf("%s/assets?ids=%s", c.baseURL, strings.Join(slugs, ","))
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add API key to headers
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch assets: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read response body for better error details
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Handle gzip decompression
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	var response CoinCapV3Response
	if err := json.NewDecoder(reader).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to our MarketData format
	var marketData []MarketData
	for _, asset := range response.Data {
		data, err := c.convertV3ToMarketData(asset)
		if err != nil {
			fmt.Printf("Error converting asset %s: %v\n", asset.Symbol, err)
			continue
		}
		marketData = append(marketData, data)
	}

	return marketData, nil
}

// convertV3ToMarketData converts CoinCapV3Asset to our MarketData format
func (c *CoinCapClient) convertV3ToMarketData(asset CoinCapV3Asset) (MarketData, error) {
	// Parse price
	price, err := strconv.ParseFloat(asset.PriceUsd, 64)
	if err != nil {
		return MarketData{}, fmt.Errorf("invalid price: %w", err)
	}

	// Parse change percent
	changePercent, err := strconv.ParseFloat(asset.ChangePercent24Hr, 64)
	if err != nil {
		changePercent = 0 // Default to 0 if parsing fails
	}

	// Calculate change amount from percentage
	changeAmount := (changePercent / 100) * price

	// Parse volume
	volume, err := strconv.ParseFloat(asset.VolumeUsd24Hr, 64)
	if err != nil {
		volume = 0
	}

	// Parse market cap
	marketCap, err := strconv.ParseFloat(asset.MarketCapUsd, 64)
	if err != nil {
		marketCap = 0
	}

	return MarketData{
		Symbol:        asset.Symbol,
		CurrentPrice:  price,
		ChangeAmount:  changeAmount,
		ChangePercent: changePercent,
		Volume:        int64(volume),
		MarketCap:     int64(marketCap),
		LastUpdated:   time.Now(),
	}, nil
}

// GetSupportedSymbols returns the list of crypto symbols we track
func (c *CoinCapClient) GetSupportedSymbols() []string {
	return []string{"BTC", "ETH", "XRP", "DOGE"}
}