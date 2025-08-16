package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewsSource represents a news source configuration
type NewsSource struct {
	ID                  uuid.UUID `json:"id"`
	Name                string    `json:"name"`
	Type                string    `json:"type"`
	URL                 string    `json:"url"`
	Language            string    `json:"language"`
	Category            string    `json:"category"`
	IsActive            bool      `json:"is_active"`
	FetchFrequencyHours int       `json:"fetch_frequency_hours"`
	LastFetched         *time.Time `json:"last_fetched"`
	CreatedAt           time.Time `json:"created_at"`
}

// NewsArticle represents a news article
type NewsArticle struct {
	ID           uuid.UUID  `json:"id"`
	SourceID     *uuid.UUID `json:"source_id"`
	Title        string     `json:"title"`
	Content      *string    `json:"content"`
	Summary      *string    `json:"summary"`
	OriginalURL  string     `json:"original_url"`
	ImageURL     *string    `json:"image_url"`
	ThumbnailURL *string    `json:"thumbnail_url"`
	FullContent  *string    `json:"full_content"`
	AuthorName   *string    `json:"author_name"`
	Keywords     *string    `json:"keywords"`
	PublishedAt  *time.Time `json:"published_at"`
	DateModified *time.Time `json:"date_modified"`
	Language     string     `json:"language"`
	Category     string     `json:"category"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	SourceName   string     `json:"source_name,omitempty"` // For joined queries
}

// UserInterest represents user interest configuration
type UserInterest struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	InterestName string    `json:"interest_name"`
	Keywords     []string  `json:"keywords"`
	Sources      []string  `json:"sources"`
	Priority     int       `json:"priority"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// GetActiveNewsSources retrieves all active news sources
func GetActiveNewsSources(db *pgxpool.Pool) ([]NewsSource, error) {
	query := `
		SELECT id, name, type, url, language, category, is_active, 
		       fetch_frequency_hours, last_fetched, created_at
		FROM news_sources 
		WHERE is_active = true
		ORDER BY category, name
	`
	
	rows, err := db.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active news sources: %w", err)
	}
	defer rows.Close()

	var sources []NewsSource
	for rows.Next() {
		var source NewsSource
		err := rows.Scan(
			&source.ID, &source.Name, &source.Type, &source.URL,
			&source.Language, &source.Category, &source.IsActive,
			&source.FetchFrequencyHours, &source.LastFetched, &source.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan news source: %w", err)
		}
		sources = append(sources, source)
	}

	return sources, nil
}

// GetNewsSourcesDueForFetch retrieves sources that need to be fetched
func GetNewsSourcesDueForFetch(db *pgxpool.Pool) ([]NewsSource, error) {
	query := `
		SELECT id, name, type, url, language, category, is_active, 
		       fetch_frequency_hours, last_fetched, created_at
		FROM news_sources 
		WHERE is_active = true 
		  AND (last_fetched IS NULL 
		       OR last_fetched < NOW() - INTERVAL '1 hour' * fetch_frequency_hours)
		ORDER BY last_fetched ASC NULLS FIRST
	`
	
	rows, err := db.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to get sources due for fetch: %w", err)
	}
	defer rows.Close()

	var sources []NewsSource
	for rows.Next() {
		var source NewsSource
		err := rows.Scan(
			&source.ID, &source.Name, &source.Type, &source.URL,
			&source.Language, &source.Category, &source.IsActive,
			&source.FetchFrequencyHours, &source.LastFetched, &source.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan news source: %w", err)
		}
		sources = append(sources, source)
	}

	return sources, nil
}

// UpdateNewsSourceLastFetched updates the last_fetched timestamp for a source
func UpdateNewsSourceLastFetched(db *pgxpool.Pool, sourceID uuid.UUID) error {
	// Get current time in Kuwait timezone
	kuwaitTZ, err := time.LoadLocation("Asia/Kuwait")
	if err != nil {
		kuwaitTZ = time.FixedZone("AST", 3*3600) // Arabia Standard Time (UTC+3)
	}
	now := time.Now().In(kuwaitTZ)
	
	query := `
		UPDATE news_sources 
		SET last_fetched = $2 
		WHERE id = $1
	`
	
	_, err = db.Exec(context.Background(), query, sourceID, now)
	if err != nil {
		return fmt.Errorf("failed to update source last_fetched: %w", err)
	}
	
	return nil
}

// CheckArticleExists checks if an article with the given URL already exists
func CheckArticleExists(db *pgxpool.Pool, originalURL string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM news_articles WHERE original_url = $1)`
	
	var exists bool
	err := db.QueryRow(context.Background(), query, originalURL).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if article exists: %w", err)
	}
	
	return exists, nil
}

// CreateNewsArticle inserts a new news article
func CreateNewsArticle(db *pgxpool.Pool, article NewsArticle) (*NewsArticle, error) {
	query := `
		INSERT INTO news_articles (source_id, title, content, summary, original_url, image_url, 
		                          thumbnail_url, full_content, author_name, keywords,
		                          published_at, date_modified, language, category)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (original_url) DO UPDATE SET
			title = EXCLUDED.title,
			content = EXCLUDED.content,
			summary = EXCLUDED.summary,
			image_url = EXCLUDED.image_url,
			thumbnail_url = EXCLUDED.thumbnail_url,
			full_content = EXCLUDED.full_content,
			author_name = EXCLUDED.author_name,
			keywords = EXCLUDED.keywords,
			date_modified = EXCLUDED.date_modified,
			updated_at = NOW()
		RETURNING id, source_id, title, content, summary, original_url, image_url,
		          thumbnail_url, full_content, author_name, keywords,
		          published_at, date_modified, language, category, created_at, updated_at
	`
	
	var created NewsArticle
	err := db.QueryRow(
		context.Background(), query,
		article.SourceID, article.Title, article.Content, article.Summary, article.OriginalURL,
		article.ImageURL, article.ThumbnailURL, article.FullContent, article.AuthorName, 
		article.Keywords, article.PublishedAt, article.DateModified, article.Language, article.Category,
	).Scan(
		&created.ID, &created.SourceID, &created.Title, &created.Content,
		&created.Summary, &created.OriginalURL, &created.ImageURL, &created.ThumbnailURL,
		&created.FullContent, &created.AuthorName, &created.Keywords,
		&created.PublishedAt, &created.DateModified, &created.Language, &created.Category,
		&created.CreatedAt, &created.UpdatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create news article: %w", err)
	}
	
	return &created, nil
}

// GetNewsArticlesWithSearchAndUserPrefs retrieves news articles with pagination, filtering, search and user preferences
func GetNewsArticlesWithSearchAndUserPrefs(db *pgxpool.Pool, userID uuid.UUID, limit, offset int, category, language, search string) ([]NewsArticle, error) {
	query := `
		SELECT na.id, na.source_id, na.title, na.content, na.summary, na.original_url,
		       na.image_url, na.thumbnail_url, na.full_content, na.author_name, na.keywords,
		       na.published_at, na.date_modified, na.language, na.category, na.created_at,
		       na.updated_at, ns.name as source_name
		FROM news_articles na
		LEFT JOIN news_sources ns ON na.source_id = ns.id
		LEFT JOIN user_news_preferences unp ON (ns.id = unp.source_id AND unp.user_id = $6)
		WHERE ($3 = '' OR na.category = $3)
		  AND ($4 = '' OR na.language = $4)
		  AND ($5 = '' OR (na.title ILIKE '%' || $5 || '%' 
		                   OR COALESCE(na.content, '') ILIKE '%' || $5 || '%'
		                   OR COALESCE(na.summary, '') ILIKE '%' || $5 || '%'
		                   OR COALESCE(na.full_content, '') ILIKE '%' || $5 || '%'))
		  AND ns.is_active = true
		  AND (unp.is_enabled IS NULL OR unp.is_enabled = true)
		ORDER BY na.published_at DESC NULLS LAST, na.created_at DESC
		LIMIT $1 OFFSET $2
	`
	
	rows, err := db.Query(context.Background(), query, limit, offset, category, language, search, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get news articles with search and user prefs: %w", err)
	}
	defer rows.Close()

	var articles []NewsArticle
	for rows.Next() {
		var article NewsArticle
		err := rows.Scan(
			&article.ID, &article.SourceID, &article.Title, &article.Content,
			&article.Summary, &article.OriginalURL, &article.ImageURL, &article.ThumbnailURL,
			&article.FullContent, &article.AuthorName, &article.Keywords,
			&article.PublishedAt, &article.DateModified, &article.Language, &article.Category,
			&article.CreatedAt, &article.UpdatedAt, &article.SourceName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan news article: %w", err)
		}
		articles = append(articles, article)
	}

	return articles, nil
}

// GetNewsArticlesWithSearch retrieves news articles with pagination, filtering and search
func GetNewsArticlesWithSearch(db *pgxpool.Pool, limit, offset int, category, language, search string) ([]NewsArticle, error) {
	query := `
		SELECT na.id, na.source_id, na.title, na.content, na.summary, na.original_url,
		       na.image_url, na.thumbnail_url, na.full_content, na.author_name, na.keywords,
		       na.published_at, na.date_modified, na.language, na.category, na.created_at,
		       na.updated_at, ns.name as source_name
		FROM news_articles na
		LEFT JOIN news_sources ns ON na.source_id = ns.id
		WHERE ($3 = '' OR na.category = $3)
		  AND ($4 = '' OR na.language = $4)
		  AND ($5 = '' OR (na.title ILIKE '%' || $5 || '%' 
		                   OR COALESCE(na.content, '') ILIKE '%' || $5 || '%'
		                   OR COALESCE(na.summary, '') ILIKE '%' || $5 || '%'
		                   OR COALESCE(na.full_content, '') ILIKE '%' || $5 || '%'))
		ORDER BY na.published_at DESC NULLS LAST, na.created_at DESC
		LIMIT $1 OFFSET $2
	`
	
	rows, err := db.Query(context.Background(), query, limit, offset, category, language, search)
	if err != nil {
		return nil, fmt.Errorf("failed to get news articles with search: %w", err)
	}
	defer rows.Close()

	var articles []NewsArticle
	for rows.Next() {
		var article NewsArticle
		err := rows.Scan(
			&article.ID, &article.SourceID, &article.Title, &article.Content,
			&article.Summary, &article.OriginalURL, &article.ImageURL, &article.ThumbnailURL,
			&article.FullContent, &article.AuthorName, &article.Keywords,
			&article.PublishedAt, &article.DateModified, &article.Language, &article.Category,
			&article.CreatedAt, &article.UpdatedAt, &article.SourceName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan news article: %w", err)
		}
		articles = append(articles, article)
	}

	return articles, nil
}

// GetNewsArticles retrieves news articles with pagination and filtering
func GetNewsArticles(db *pgxpool.Pool, limit, offset int, category, language string) ([]NewsArticle, error) {
	query := `
		SELECT na.id, na.source_id, na.title, na.content, na.summary, na.original_url,
		       na.image_url, na.thumbnail_url, na.full_content, na.author_name, na.keywords,
		       na.published_at, na.date_modified, na.language, na.category, na.created_at,
		       na.updated_at, ns.name as source_name
		FROM news_articles na
		LEFT JOIN news_sources ns ON na.source_id = ns.id
		WHERE ($3 = '' OR na.category = $3)
		  AND ($4 = '' OR na.language = $4)
		ORDER BY na.published_at DESC NULLS LAST, na.created_at DESC
		LIMIT $1 OFFSET $2
	`
	
	rows, err := db.Query(context.Background(), query, limit, offset, category, language)
	if err != nil {
		return nil, fmt.Errorf("failed to get news articles: %w", err)
	}
	defer rows.Close()

	var articles []NewsArticle
	for rows.Next() {
		var article NewsArticle
		err := rows.Scan(
			&article.ID, &article.SourceID, &article.Title, &article.Content,
			&article.Summary, &article.OriginalURL, &article.ImageURL, &article.ThumbnailURL,
			&article.FullContent, &article.AuthorName, &article.Keywords,
			&article.PublishedAt, &article.DateModified, &article.Language, &article.Category,
			&article.CreatedAt, &article.UpdatedAt, &article.SourceName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan news article: %w", err)
		}
		articles = append(articles, article)
	}

	return articles, nil
}

// GetNewsArticleByID retrieves a single news article by ID
func GetNewsArticleByID(db *pgxpool.Pool, articleID uuid.UUID) (*NewsArticle, error) {
	query := `
		SELECT na.id, na.source_id, na.title, na.content, na.summary, na.original_url,
		       na.image_url, na.thumbnail_url, na.full_content, na.author_name, na.keywords,
		       na.published_at, na.date_modified, na.language, na.category, na.created_at,
		       na.updated_at, ns.name as source_name
		FROM news_articles na
		LEFT JOIN news_sources ns ON na.source_id = ns.id
		WHERE na.id = $1
	`
	
	var article NewsArticle
	err := db.QueryRow(context.Background(), query, articleID).Scan(
		&article.ID, &article.SourceID, &article.Title, &article.Content,
		&article.Summary, &article.OriginalURL, &article.ImageURL, &article.ThumbnailURL,
		&article.FullContent, &article.AuthorName, &article.Keywords,
		&article.PublishedAt, &article.DateModified, &article.Language, &article.Category,
		&article.CreatedAt, &article.UpdatedAt, &article.SourceName,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get news article by ID: %w", err)
	}
	
	return &article, nil
}

// GetNewsArticleCountWithSearchAndUserPrefs returns the total count of articles with filters, search and user preferences
func GetNewsArticleCountWithSearchAndUserPrefs(db *pgxpool.Pool, userID uuid.UUID, category, language, search string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM news_articles na
		LEFT JOIN news_sources ns ON na.source_id = ns.id
		LEFT JOIN user_news_preferences unp ON (ns.id = unp.source_id AND unp.user_id = $4)
		WHERE ($1 = '' OR na.category = $1)
		  AND ($2 = '' OR na.language = $2)
		  AND ($3 = '' OR (na.title ILIKE '%' || $3 || '%' 
		                   OR COALESCE(na.content, '') ILIKE '%' || $3 || '%'
		                   OR COALESCE(na.summary, '') ILIKE '%' || $3 || '%'
		                   OR COALESCE(na.full_content, '') ILIKE '%' || $3 || '%'))
		  AND ns.is_active = true
		  AND (unp.is_enabled IS NULL OR unp.is_enabled = true)
	`
	
	var count int
	err := db.QueryRow(context.Background(), query, category, language, search, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get news article count with search and user prefs: %w", err)
	}
	
	return count, nil
}

// GetNewsArticleCountWithSearch returns the total count of articles with filters and search
func GetNewsArticleCountWithSearch(db *pgxpool.Pool, category, language, search string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM news_articles
		WHERE ($1 = '' OR category = $1)
		  AND ($2 = '' OR language = $2)
		  AND ($3 = '' OR (title ILIKE '%' || $3 || '%' 
		                   OR COALESCE(content, '') ILIKE '%' || $3 || '%'
		                   OR COALESCE(summary, '') ILIKE '%' || $3 || '%'
		                   OR COALESCE(full_content, '') ILIKE '%' || $3 || '%'))
	`
	
	var count int
	err := db.QueryRow(context.Background(), query, category, language, search).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get news article count with search: %w", err)
	}
	
	return count, nil
}

// GetNewsArticleCount returns the total count of articles with filters
func GetNewsArticleCount(db *pgxpool.Pool, category, language string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM news_articles
		WHERE ($1 = '' OR category = $1)
		  AND ($2 = '' OR language = $2)
	`
	
	var count int
	err := db.QueryRow(context.Background(), query, category, language).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get news article count: %w", err)
	}
	
	return count, nil
}

// UserNewsPreference represents a user's preference for a news source
type UserNewsPreference struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	SourceID  uuid.UUID `json:"source_id"`
	IsEnabled bool      `json:"is_enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewsSourceWithUserPref combines news source with user preference
type NewsSourceWithUserPref struct {
	NewsSource
	UserEnabled bool `json:"user_enabled"`
}

// GetUserNewsPreferences gets all user preferences for news sources
func GetUserNewsPreferences(db *pgxpool.Pool, userID uuid.UUID) (map[uuid.UUID]bool, error) {
	query := `SELECT source_id, is_enabled FROM user_news_preferences WHERE user_id = $1`
	
	rows, err := db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user news preferences: %w", err)
	}
	defer rows.Close()

	preferences := make(map[uuid.UUID]bool)
	for rows.Next() {
		var sourceID uuid.UUID
		var isEnabled bool
		err := rows.Scan(&sourceID, &isEnabled)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user preference: %w", err)
		}
		preferences[sourceID] = isEnabled
	}

	return preferences, nil
}

// GetAllNewsSources retrieves all news sources
func GetAllNewsSources(db *pgxpool.Pool) ([]NewsSource, error) {
	query := `
		SELECT id, name, type, url, language, category, is_active, 
		       fetch_frequency_hours, last_fetched, created_at
		FROM news_sources 
		ORDER BY category, name
	`
	
	rows, err := db.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all news sources: %w", err)
	}
	defer rows.Close()

	var sources []NewsSource
	for rows.Next() {
		var source NewsSource
		err := rows.Scan(
			&source.ID, &source.Name, &source.Type, &source.URL,
			&source.Language, &source.Category, &source.IsActive,
			&source.FetchFrequencyHours, &source.LastFetched, &source.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan news source: %w", err)
		}
		sources = append(sources, source)
	}

	return sources, nil
}

// GetNewsSourcesWithUserPrefs gets news sources with user preferences included
func GetNewsSourcesWithUserPrefs(db *pgxpool.Pool, userID uuid.UUID) ([]NewsSourceWithUserPref, error) {
	// Get all news sources (not just active ones)
	sources, err := GetAllNewsSources(db)
	if err != nil {
		return nil, err
	}

	// Get user preferences
	userPrefs, err := GetUserNewsPreferences(db, userID)
	if err != nil {
		return nil, err
	}

	// Combine sources with user preferences
	sourcesWithPrefs := make([]NewsSourceWithUserPref, len(sources))
	for i, source := range sources {
		userEnabled, exists := userPrefs[source.ID]
		if !exists {
			// Default to enabled if no preference exists
			userEnabled = true
		}
		
		sourcesWithPrefs[i] = NewsSourceWithUserPref{
			NewsSource:  source,
			UserEnabled: userEnabled,
		}
	}

	return sourcesWithPrefs, nil
}

// SetUserNewsPreference sets or updates a user's preference for a news source
func SetUserNewsPreference(db *pgxpool.Pool, userID, sourceID uuid.UUID, isEnabled bool) error {
	query := `
		INSERT INTO user_news_preferences (user_id, source_id, is_enabled)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, source_id)
		DO UPDATE SET is_enabled = $3, updated_at = CURRENT_TIMESTAMP
	`
	
	_, err := db.Exec(context.Background(), query, userID, sourceID, isEnabled)
	if err != nil {
		return fmt.Errorf("failed to set user news preference: %w", err)
	}
	
	return nil
}

// ToggleNewsSourceActive toggles the active status of a news source
func ToggleNewsSourceActive(db *pgxpool.Pool, sourceID uuid.UUID) (*NewsSource, error) {
	query := `
		UPDATE news_sources 
		SET is_active = NOT is_active
		WHERE id = $1
		RETURNING id, name, url, type, language, category, is_active, 
		          fetch_frequency_hours, last_fetched, created_at
	`
	
	var source NewsSource
	err := db.QueryRow(context.Background(), query, sourceID).Scan(
		&source.ID, &source.Name, &source.URL, &source.Type,
		&source.Language, &source.Category, &source.IsActive,
		&source.FetchFrequencyHours, &source.LastFetched,
		&source.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to toggle news source active status: %w", err)
	}
	
	return &source, nil
}

// DeleteOldNewsArticles removes articles older than specified days
func DeleteOldNewsArticles(db *pgxpool.Pool, days int) (int, error) {
	query := `
		DELETE FROM news_articles 
		WHERE created_at < NOW() - INTERVAL '%d days'
	`
	
	result, err := db.Exec(context.Background(), fmt.Sprintf(query, days))
	if err != nil {
		return 0, fmt.Errorf("failed to delete old news articles: %w", err)
	}
	
	return int(result.RowsAffected()), nil
}

// GetUserInterests retrieves user interests
func GetUserInterests(db *pgxpool.Pool, userID uuid.UUID) ([]UserInterest, error) {
	query := `
		SELECT id, user_id, interest_name, keywords, sources, priority, is_active, created_at
		FROM user_interests
		WHERE user_id = $1 AND is_active = true
		ORDER BY priority DESC, interest_name
	`
	
	rows, err := db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user interests: %w", err)
	}
	defer rows.Close()

	var interests []UserInterest
	for rows.Next() {
		var interest UserInterest
		var keywordsJSON, sourcesJSON []byte
		
		err := rows.Scan(
			&interest.ID, &interest.UserID, &interest.InterestName,
			&keywordsJSON, &sourcesJSON, &interest.Priority,
			&interest.IsActive, &interest.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user interest: %w", err)
		}
		
		// Parse JSON fields
		if len(keywordsJSON) > 0 {
			// Simple JSON parsing for string arrays
			// TODO: Use proper JSON unmarshaling if needed
		}
		
		interests = append(interests, interest)
	}

	return interests, nil
}

// GetActiveSourceCountByCategory returns the count of active sources for a specific category
func GetActiveSourceCountByCategory(db *pgxpool.Pool, category string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM news_sources 
		WHERE is_active = true AND category = $1
	`
	
	var count int
	err := db.QueryRow(context.Background(), query, category).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get active source count for category %s: %w", category, err)
	}
	
	return count, nil
}

// GetActiveSourceCountByCategoryForUser returns the count of sources that are active AND enabled by the user
func GetActiveSourceCountByCategoryForUser(db *pgxpool.Pool, userID uuid.UUID, category string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM news_sources ns
		LEFT JOIN user_news_preferences unp ON (ns.id = unp.source_id AND unp.user_id = $1)
		WHERE ns.is_active = true 
		  AND ns.category = $2
		  AND (unp.is_enabled IS NULL OR unp.is_enabled = true)
	`
	
	var count int
	err := db.QueryRow(context.Background(), query, userID, category).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get user-enabled source count for category %s: %w", category, err)
	}
	
	return count, nil
}