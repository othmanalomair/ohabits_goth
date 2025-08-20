package market

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"ohabits.com/internal/db"
)

// FetchService handles market data fetching operations
type FetchService struct {
	db           *pgxpool.Pool
	cryptoClient *CoinCapClient
	yahooClient  *YahooFinanceClient
	isRunning    bool
	stopChan     chan bool
}

// NewFetchService creates a new market data fetch service
func NewFetchService(database *pgxpool.Pool) *FetchService {
	return &FetchService{
		db:           database,
		cryptoClient: NewCoinCapClient(),
		yahooClient:  NewYahooFinanceClient(),
		stopChan:     make(chan bool, 1),
	}
}

// StartBackgroundFetching starts the background market data fetching process
func (s *FetchService) StartBackgroundFetching() {
	if s.isRunning {
		log.Println("Market data fetching service is already running")
		return
	}

	s.isRunning = true
	log.Println("Starting market data fetching background service...")

	// Run immediately on start
	go s.fetchMarketData()

	// Set up ticker for periodic fetching (every 5 minutes for crypto prices)
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				if !s.isRunning {
					ticker.Stop()
					return
				}
				s.fetchMarketData()
			case <-s.stopChan:
				ticker.Stop()
				return
			}
		}
	}()

	log.Println("Market data fetching service started successfully")
}

// StopBackgroundFetching stops the background fetching process
func (s *FetchService) StopBackgroundFetching() {
	if !s.isRunning {
		return
	}

	log.Println("Stopping market data fetching service...")
	s.isRunning = false
	
	// Send stop signal
	select {
	case s.stopChan <- true:
	default:
	}

	log.Println("Market data fetching service stopped")
}

// fetchMarketData fetches current market data and updates the database
func (s *FetchService) fetchMarketData() {
	log.Println("Fetching market data from multiple sources...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var allMarketData []MarketData
	totalUpdated := 0

	// Fetch cryptocurrency data from CoinCap
	log.Println("Fetching crypto data from CoinCap...")
	cryptoSymbols := s.cryptoClient.GetSupportedSymbols()
	cryptoData, err := s.cryptoClient.GetAssets(cryptoSymbols)
	if err != nil {
		log.Printf("Error fetching crypto data: %v", err)
	} else {
		allMarketData = append(allMarketData, cryptoData...)
		log.Printf("Fetched %d crypto assets", len(cryptoData))
	}

	// Fetch traditional market data from Yahoo Finance
	log.Println("Fetching market data from Yahoo Finance...")
	tickers := s.yahooClient.GetSupportedTickers()
	yahooData, err := s.yahooClient.GetAssets(tickers)
	if err != nil {
		log.Printf("Error fetching Yahoo Finance data: %v", err)
	} else {
		allMarketData = append(allMarketData, yahooData...)
		log.Printf("Fetched %d traditional market assets", len(yahooData))
	}

	// Update database with all market data
	for _, data := range allMarketData {
		dbData := db.MarketData{
			Symbol:        data.Symbol,
			CurrentPrice:  data.CurrentPrice,
			ChangeAmount:  data.ChangeAmount,
			ChangePercent: data.ChangePercent,
			Volume:        data.Volume,
			MarketCap:     data.MarketCap,
			LastUpdated:   data.LastUpdated,
		}
		
		if err := db.UpdateMarketData(ctx, s.db, dbData); err != nil {
			log.Printf("Error updating market data for %s: %v", data.Symbol, err)
		} else {
			totalUpdated++
		}
	}

	log.Printf("Market data fetch completed: %d/%d symbols updated", totalUpdated, len(allMarketData))
}

// FetchNow triggers an immediate fetch of market data
func (s *FetchService) FetchNow() error {
	log.Println("Manual market data fetch triggered")
	s.fetchMarketData()
	return nil
}