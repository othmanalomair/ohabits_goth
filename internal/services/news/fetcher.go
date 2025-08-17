package news

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"ohabits.com/internal/db"
)

// FetchService handles news fetching operations
type FetchService struct {
	db        *pgxpool.Pool
	parser    *RSSParser
	scraper   *ArticleScraper
	isRunning bool
}

// NewFetchService creates a new news fetch service
func NewFetchService(database *pgxpool.Pool) *FetchService {
	return &FetchService{
		db:      database,
		parser:  NewRSSParser(),
		scraper: NewArticleScraper(),
	}
}

// StartBackgroundFetching starts the background news fetching process
func (s *FetchService) StartBackgroundFetching() {
	if s.isRunning {
		log.Println("News fetching service is already running")
		return
	}

	s.isRunning = true
	log.Println("Starting news fetching background service...")

	// Run immediately on start
	go s.fetchAllDueSources()

	// Set up ticker for periodic fetching (every 30 minutes)
	ticker := time.NewTicker(30 * time.Minute)
	go func() {
		for range ticker.C {
			if !s.isRunning {
				ticker.Stop()
				return
			}
			s.fetchAllDueSources()
		}
	}()
}

// StopBackgroundFetching stops the background news fetching process
func (s *FetchService) StopBackgroundFetching() {
	s.isRunning = false
	log.Println("Stopped news fetching background service")
}

// fetchAllDueSources fetches news from all sources that are due for updates
func (s *FetchService) fetchAllDueSources() {
	log.Println("Checking for news sources due for fetching...")

	sources, err := db.GetNewsSourcesDueForFetch(s.db)
	if err != nil {
		log.Printf("Error getting sources due for fetch: %v", err)
		return
	}

	if len(sources) == 0 {
		log.Println("No news sources due for fetching")
		return
	}

	log.Printf("Found %d news sources due for fetching", len(sources))

	for _, source := range sources {
		s.fetchFromSource(source)
		// Add delay between sources to be respectful, especially for Reddit
		if strings.Contains(source.URL, "reddit.com") {
			// Longer delay for Reddit due to strict rate limiting
			time.Sleep(5 * time.Second)
		} else {
			// Standard delay for other sources
			time.Sleep(2 * time.Second)
		}
	}
}

// fetchFromSource fetches news from a specific source
func (s *FetchService) fetchFromSource(source db.NewsSource) {
	log.Printf("Fetching news from %s (%s)...", source.Name, source.URL)

	// Update last_fetched timestamp immediately to prevent duplicate fetching
	err := db.UpdateNewsSourceLastFetched(s.db, source.ID)
	if err != nil {
		log.Printf("Error updating last_fetched for %s: %v", source.Name, err)
	}

	// Parse RSS feed
	articles, err := s.parser.ParseRSSFeed(source.URL)
	if err != nil {
		log.Printf("Error parsing RSS feed for %s: %v", source.Name, err)
		return
	}

	if len(articles) == 0 {
		log.Printf("No articles found in RSS feed for %s", source.Name)
		return
	}

	// Save articles to database
	savedCount := 0
	skippedCount := 0
	for _, article := range articles {
		// Check if article already exists
		exists, err := db.CheckArticleExists(s.db, article.URL)
		if err != nil {
			log.Printf("Error checking if article exists '%s': %v", article.Title, err)
			continue
		}
		
		if exists {
			skippedCount++
			continue // Skip existing articles
		}

		// Try to scrape full article content if it's from supported Kuwait sources
		var enhancedArticle *EnhancedArticle
		if s.scraper.IsKuwaitNewsSource(article.URL) {
			enhancedArticle, err = s.scraper.ScrapeFullArticle(article.URL)
			if err != nil {
				log.Printf("Error scraping full content for '%s': %v", article.Title, err)
				// Continue with RSS data if scraping fails
			}
		}

		// Create database article with enhanced data if available
		dbArticle := db.NewsArticle{
			SourceID:    &source.ID,
			OriginalURL: article.URL,
			Language:    source.Language,
			Category:    source.Category,
		}

		if enhancedArticle != nil {
			// Use enhanced data from scraping
			dbArticle.Title = enhancedArticle.Title
			dbArticle.Content = &enhancedArticle.Summary
			dbArticle.Summary = stringToPointer(enhancedArticle.Summary)
			dbArticle.FullContent = &enhancedArticle.FullContent
			dbArticle.ImageURL = stringToPointer(enhancedArticle.ImageURL)
			dbArticle.ThumbnailURL = stringToPointer(enhancedArticle.ThumbnailURL)
			dbArticle.AuthorName = stringToPointer(enhancedArticle.AuthorName)
			dbArticle.Keywords = stringToPointer(enhancedArticle.Keywords)
			dbArticle.PublishedAt = timeToPointer(enhancedArticle.PublishedAt)
			dbArticle.DateModified = timeToPointer(enhancedArticle.ModifiedAt)
		} else {
			// Use RSS data as fallback
			dbArticle.Title = article.Title
			dbArticle.Content = &article.Content
			dbArticle.Summary = stringToPointer(article.Content)
			dbArticle.ImageURL = stringToPointer(article.ImageURL)
			dbArticle.PublishedAt = timeToPointer(article.PublishedAt)
		}

		_, err = db.CreateNewsArticle(s.db, dbArticle)
		if err != nil {
			log.Printf("Error saving article '%s': %v", article.Title, err)
			continue
		}
		savedCount++
	}

	log.Printf("Successfully saved %d/%d articles from %s (skipped %d duplicates)", savedCount, len(articles), source.Name, skippedCount)
}

// FetchFromSourceManually manually fetches news from a specific source
func (s *FetchService) FetchFromSourceManually(sourceID uuid.UUID) error {
	// Get source details
	sources, err := db.GetActiveNewsSources(s.db)
	if err != nil {
		return fmt.Errorf("failed to get news sources: %w", err)
	}

	var targetSource *db.NewsSource
	for _, source := range sources {
		if source.ID == sourceID {
			targetSource = &source
			break
		}
	}

	if targetSource == nil {
		return fmt.Errorf("news source not found: %s", sourceID)
	}

	s.fetchFromSource(*targetSource)
	return nil
}

// GetNewsSourcesStatus returns the status of all news sources
func (s *FetchService) GetNewsSourcesStatus() ([]db.NewsSource, error) {
	return db.GetActiveNewsSources(s.db)
}

// CleanupOldArticles removes articles older than specified days
func (s *FetchService) CleanupOldArticles(days int) (int, error) {
	deletedCount, err := db.DeleteOldNewsArticles(s.db, days)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old articles: %w", err)
	}

	log.Printf("Cleaned up %d old news articles (older than %d days)", deletedCount, days)
	return deletedCount, nil
}

// Utility functions
func stringToPointer(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func timeToPointer(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}