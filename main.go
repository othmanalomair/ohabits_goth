package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"ohabits.com/cmd/server"
	"ohabits.com/internal/db"
	"ohabits.com/internal/services"
	"ohabits.com/internal/services/news"
	"ohabits.com/internal/services/market"
)

func main() {
	db.Connect()
	defer db.Close()

	// Start the sync service for automatic episode updates
	syncService := services.NewSyncService(db.DB)
	syncService.Start()
	defer syncService.Stop()

	// Start the news fetching service
	newsService := news.NewFetchService(db.DB)
	newsService.StartBackgroundFetching()
	defer newsService.StopBackgroundFetching()

	// Start the market data fetching service
	marketService := market.NewFetchService(db.DB)
	marketService.StartBackgroundFetching()
	defer marketService.StopBackgroundFetching()

	// Start the finance auto-payment service
	financeService := services.NewFinanceService(db.DB)
	financeService.StartBackgroundProcessing()
	defer financeService.StopBackgroundProcessing()

	srv := server.Server() // Now returns an *http.Server

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		log.Println("Server is running on port 8080...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down the server...")
	syncService.Stop()
	newsService.StopBackgroundFetching()
	marketService.StopBackgroundFetching()
	financeService.StopBackgroundProcessing()
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatalf("Server Shutdown: %v", err)
	}
	log.Println("Server stopped.")
}
