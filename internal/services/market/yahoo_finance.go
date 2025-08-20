package market

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// YahooFinanceResponse represents the response from Yahoo Finance API
type YahooFinanceResponse struct {
	Chart YahooChart `json:"chart"`
}

type YahooChart struct {
	Result []YahooResult `json:"result"`
	Error  interface{}   `json:"error"`
}

type YahooResult struct {
	Meta       YahooMeta       `json:"meta"`
	Timestamp  []int64         `json:"timestamp"`
	Indicators YahooIndicators `json:"indicators"`
}

type YahooMeta struct {
	Currency                 string  `json:"currency"`
	Symbol                   string  `json:"symbol"`
	ExchangeName             string  `json:"exchangeName"`
	FullExchangeName         string  `json:"fullExchangeName"`
	InstrumentType           string  `json:"instrumentType"`
	RegularMarketTime        int64   `json:"regularMarketTime"`
	RegularMarketPrice       float64 `json:"regularMarketPrice"`
	FiftyTwoWeekHigh         float64 `json:"fiftyTwoWeekHigh"`
	FiftyTwoWeekLow          float64 `json:"fiftyTwoWeekLow"`
	RegularMarketDayHigh     float64 `json:"regularMarketDayHigh"`
	RegularMarketDayLow      float64 `json:"regularMarketDayLow"`
	RegularMarketVolume      int64   `json:"regularMarketVolume"`
	LongName                 string  `json:"longName"`
	ShortName                string  `json:"shortName"`
	ChartPreviousClose       float64 `json:"chartPreviousClose"`
}

type YahooIndicators struct {
	Quote    []YahooQuote    `json:"quote"`
	AdjClose []YahooAdjClose `json:"adjclose"`
}

type YahooQuote struct {
	Close  []float64 `json:"close"`
	High   []float64 `json:"high"`
	Volume []int64   `json:"volume"`
	Open   []float64 `json:"open"`
	Low    []float64 `json:"low"`
}

type YahooAdjClose struct {
	AdjClose []float64 `json:"adjclose"`
}

// YahooFinanceClient handles communication with the Yahoo Finance API
type YahooFinanceClient struct {
	client  *http.Client
	baseURL string
}

// NewYahooFinanceClient creates a new Yahoo Finance API client
func NewYahooFinanceClient() *YahooFinanceClient {
	return &YahooFinanceClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://query1.finance.yahoo.com/v8/finance/chart",
	}
}

// GetAsset fetches market data for a specific ticker symbol
func (c *YahooFinanceClient) GetAsset(ticker string) (*MarketData, error) {
	url := fmt.Sprintf("%s/%s?interval=1d&range=1d", c.baseURL, ticker)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ohabits/1.0)")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var response YahooFinanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data returned for ticker %s", ticker)
	}

	return c.convertToMarketData(response.Chart.Result[0], ticker)
}

// GetAssets fetches market data for multiple ticker symbols
func (c *YahooFinanceClient) GetAssets(tickers []string) ([]MarketData, error) {
	var marketData []MarketData
	
	for _, ticker := range tickers {
		data, err := c.GetAsset(ticker)
		if err != nil {
			fmt.Printf("Error fetching data for %s: %v\n", ticker, err)
			continue
		}
		marketData = append(marketData, *data)
	}
	
	return marketData, nil
}

// convertToMarketData converts Yahoo Finance data to our MarketData format
func (c *YahooFinanceClient) convertToMarketData(result YahooResult, ticker string) (*MarketData, error) {
	meta := result.Meta
	
	// Calculate change percentage
	var changePercent float64
	var changeAmount float64
	
	if meta.ChartPreviousClose > 0 {
		changeAmount = meta.RegularMarketPrice - meta.ChartPreviousClose
		changePercent = (changeAmount / meta.ChartPreviousClose) * 100
	}

	// Convert ticker to display symbol
	symbol := c.tickerToSymbol(ticker)
	
	return &MarketData{
		Symbol:        symbol,
		CurrentPrice:  meta.RegularMarketPrice,
		ChangeAmount:  changeAmount,
		ChangePercent: changePercent,
		Volume:        meta.RegularMarketVolume,
		MarketCap:     0, // Not available in Yahoo Finance for all instruments
		LastUpdated:   time.Now(),
	}, nil
}

// tickerToSymbol converts Yahoo Finance tickers to our display symbols
func (c *YahooFinanceClient) tickerToSymbol(ticker string) string {
	tickerMap := map[string]string{
		"GLD":      "GOLD",
		"CL=F":     "OIL",
		"^GSPC":    "SPX",
		"BOURSA.KW": "BOURSA",
	}
	
	if symbol, exists := tickerMap[ticker]; exists {
		return symbol
	}
	return ticker
}

// GetSupportedTickers returns the list of tickers we track
func (c *YahooFinanceClient) GetSupportedTickers() []string {
	return []string{"GLD", "CL=F", "^GSPC", "BOURSA.KW"}
}

// GetSupportedAssets returns the list of assets with their display names
func (c *YahooFinanceClient) GetSupportedAssets() []struct {
	Ticker string
	Symbol string
	Name   string
	Type   string
} {
	return []struct {
		Ticker string
		Symbol string
		Name   string
		Type   string
	}{
		{"GLD", "GOLD", "Gold ETF", "commodity"},
		{"CL=F", "OIL", "Crude Oil", "commodity"},
		{"^GSPC", "SPX", "S&P 500", "index"},
		{"BOURSA.KW", "BOURSA", "Boursa Kuwait", "stock"},
	}
}