package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"ohabits.com/internal/db"
	"ohabits.com/internal/services/news"
	"time"
)

// TemplateArticle represents a news article for template rendering
type TemplateArticle struct {
	ID           string
	Title        string
	Content      string
	Summary      string
	FullContent  string
	OriginalURL  string
	ImageURL     string
	ThumbnailURL string
	AuthorName   string
	Keywords     string
	PublishedAt  string
	ModifiedAt   string
	Language     string
	Category     string
	SourceName   string
}

// NewsPageData represents data for the news page
type NewsPageData struct {
	User         db.User
	Articles     []TemplateArticle
	CurrentPage  int
	TotalPages   int
	TotalCount   int
	Category     string
	Language     string
	Search       string
	Sources      []db.NewsSource
	HasNext      bool
	HasPrev      bool
	NextPage     int
	PrevPage     int
}

// convertToTemplateArticles converts database articles to template-friendly articles
func convertToTemplateArticles(articles []db.NewsArticle) []TemplateArticle {
	templateArticles := make([]TemplateArticle, len(articles))
	for i, article := range articles {
		templateArticles[i] = TemplateArticle{
			ID:           article.ID.String(),
			Title:        article.Title,
			Content:      stringPtrToString(article.Content),
			Summary:      stringPtrToString(article.Summary),
			FullContent:  stringPtrToString(article.FullContent),
			OriginalURL:  article.OriginalURL,
			ImageURL:     stringPtrToString(article.ImageURL),
			ThumbnailURL: stringPtrToString(article.ThumbnailURL),
			AuthorName:   stringPtrToString(article.AuthorName),
			Keywords:     stringPtrToString(article.Keywords),
			PublishedAt:  timePtrToString(article.PublishedAt),
			ModifiedAt:   timePtrToString(article.DateModified),
			Language:     article.Language,
			Category:     article.Category,
			SourceName:   article.SourceName,
		}
	}
	return templateArticles
}

// Helper functions to safely convert pointers to strings
func stringPtrToString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func timePtrToString(ptr *time.Time) string {
	if ptr == nil {
		return ""
	}
	
	// Since published_at is stored as "timestamp without time zone" in the database,
	// the time is already in Kuwait timezone but Go treats it as local system time.
	// We need to interpret it as Kuwait time and display it as-is.
	
	// Check if this timestamp has no timezone info (from database)
	// In this case, we assume it's already in Kuwait timezone
	if ptr.Location() == time.UTC || ptr.Location().String() == "Local" {
		// This timestamp is from the database and should be treated as Kuwait time
		// Just format it directly without conversion
		return ptr.Format("Jan 2, 2006 15:04")
	}
	
	// If it has timezone info, check if it's already Kuwait timezone
	_, offset := ptr.Zone()
	if offset == 3*3600 {
		// Already in Kuwait timezone, don't convert again
		return ptr.Format("Jan 2, 2006 15:04")
	}
	
	// Convert to Kuwait timezone for display only if it has different timezone
	kuwaitTZ, err := time.LoadLocation("Asia/Kuwait")
	if err != nil {
		// Fallback to UTC+3 if Kuwait timezone is not available
		kuwaitTZ = time.FixedZone("AST", 3*3600) // Arabia Standard Time (UTC+3)
	}
	
	return ptr.In(kuwaitTZ).Format("Jan 2, 2006 15:04")
}

// NewsHandler displays the news page
func NewsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("NewsHandler called with method: %s, URL: %s, HTMX: %s\n", r.Method, r.URL.String(), r.Header.Get("HX-Request"))
	// Get user from context
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		// Check if this is an HTMX request
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Redirect", "/login")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		// Check if this is an HTMX request
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Redirect", "/login")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := db.GetUser(db.DB, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	category := r.URL.Query().Get("category")
	language := r.URL.Query().Get("language")
	search := r.URL.Query().Get("search")
	
	// Default to Arabic/Kuwait news if no language specified
	if language == "" {
		language = "ar"
	}
	if category == "" && language == "ar" {
		category = "kuwait"
	}

	const articlesPerPage = 20
	offset := (page - 1) * articlesPerPage

	// Get articles from database with user preferences
	var articles []db.NewsArticle
	var totalCount int
	if search != "" {
		articles, err = db.GetNewsArticlesWithSearchAndUserPrefs(db.DB, userID, articlesPerPage, offset, category, language, search)
		if err != nil {
			http.Error(w, "Failed to search news articles", http.StatusInternalServerError)
			fmt.Printf("Error searching news articles: %v\n", err)
			return
		}
		
		totalCount, err = db.GetNewsArticleCountWithSearchAndUserPrefs(db.DB, userID, category, language, search)
	} else {
		articles, err = db.GetNewsArticlesWithSearchAndUserPrefs(db.DB, userID, articlesPerPage, offset, category, language, "")
		if err != nil {
			http.Error(w, "Failed to load news articles", http.StatusInternalServerError)
			fmt.Printf("Error loading news articles: %v\n", err)
			return
		}
		
		totalCount, err = db.GetNewsArticleCountWithSearchAndUserPrefs(db.DB, userID, category, language, "")
	}
	if err != nil {
		http.Error(w, "Failed to count news articles", http.StatusInternalServerError)
		fmt.Printf("Error counting news articles: %v\n", err)
		return
	}

	// Get news sources for filter
	sources, err := db.GetActiveNewsSources(db.DB)
	if err != nil {
		fmt.Printf("Error loading news sources: %v\n", err)
		sources = []db.NewsSource{} // Continue with empty sources
	}

	// Calculate pagination
	totalPages := (totalCount + articlesPerPage - 1) / articlesPerPage
	if totalPages == 0 {
		totalPages = 1
	}

	data := NewsPageData{
		User:        user,
		Articles:    convertToTemplateArticles(articles),
		CurrentPage: page,
		TotalPages:  totalPages,
		TotalCount:  totalCount,
		Category:    category,
		Language:    language,
		Search:      search,
		Sources:     sources,
		HasNext:     page < totalPages,
		HasPrev:     page > 1,
		NextPage:    page + 1,
		PrevPage:    page - 1,
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		// Return just the article content for HTMX requests
		fmt.Printf("HTMX request detected for page %d\n", page)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		
		// Use enhanced template parsing for HTMX requests
		tmpl, err := template.New("news_articles.html").Funcs(template.FuncMap{
			"substr": func(s string, start, length int) string {
				if start < 0 {
					start = 0
				}
				// Convert to runes to handle UTF-8 properly
				runes := []rune(s)
				if start >= len(runes) {
					return ""
				}
				end := start + length
				if end > len(runes) {
					end = len(runes)
				}
				return string(runes[start:end])
			},
		}).ParseFiles("templates/partials/news_articles.html")
		if err != nil {
			fmt.Printf("HTMX template parsing error: %v\n", err)
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			fmt.Printf("HTMX template execution error: %v\n", err)
			http.Error(w, "Template execution error", http.StatusInternalServerError)
		}
		fmt.Printf("HTMX response sent with %d articles\n", len(articles))
		return
	}

	// Render full page
	tmpl, err := template.New("base.html").Funcs(template.FuncMap{
		"substr": func(s string, start, length int) string {
			if start < 0 {
				start = 0
			}
			// Convert to runes to handle UTF-8 properly
			runes := []rune(s)
			if start >= len(runes) {
				return ""
			}
			end := start + length
			if end > len(runes) {
				end = len(runes)
			}
			return string(runes[start:end])
		},
	}).ParseFiles("templates/base.html", "templates/news.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		fmt.Printf("Template parsing error: %v\n", err)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		fmt.Printf("Template execution error: %v\n", err)
	}
}

// renderNewsArticles renders just the articles list for HTMX updates
func renderNewsArticles(w http.ResponseWriter, data NewsPageData) {
	fmt.Printf("renderNewsArticles called with %d articles, page %d of %d\n", len(data.Articles), data.CurrentPage, data.TotalPages)
	
	// Add error recovery
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in renderNewsArticles: %v\n", r)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}()
	
	// Render manually to avoid template issues
	fmt.Fprintf(w, `<div class="news-stats">
		<span class="article-count">%d articles found</span>
	</div>
	
	<div class="articles-grid">`, data.TotalCount)
	
	for _, article := range data.Articles {
		content := article.Content
		// Handle UTF-8 properly for content truncation
		if len([]rune(content)) > 200 {
			runes := []rune(content)
			content = string(runes[:200]) + "..."
		}
		
		fmt.Fprintf(w, `
		<article class="news-article %s">
			<div class="article-content">
				<div class="article-meta">
					<span class="article-source">%s</span>
					<span class="article-category">%s</span>
					<time class="article-date">%s</time>
				</div>
				
				<h2 class="article-title">
					<a href="%s" target="_blank" rel="noopener noreferrer">%s</a>
				</h2>
				
				<div class="article-summary">%s</div>
				
				<div class="article-actions">
					<a href="%s" target="_blank" rel="noopener noreferrer" class="read-more-btn">
						%s
						<img src="/static/images/svg/arrow-right-circle.svg" alt="Read more" />
					</a>
				</div>
			</div>
		</article>`,
			func() string { if article.Language == "ar" { return "rtl" }; return "" }(),
			article.SourceName,
			article.Category,
			article.PublishedAt,
			article.OriginalURL,
			article.Title,
			content,
			article.OriginalURL,
			func() string { if article.Language == "ar" { return "اقرأ المزيد" }; return "Read More" }(),
		)
	}
	
	fmt.Fprintf(w, `</div>
	
	<!-- Pagination -->
	<div class="news-pagination">`)
	
	if data.HasPrev {
		fmt.Fprintf(w, `
		<button class="pagination-btn" 
				hx-get="/news?page=%d&category=%s&language=%s"
				hx-target="#news-content" 
				hx-swap="innerHTML">
			<img src="/static/images/svg/arrow-left-circle.svg" alt="Previous" />
			Previous
		</button>`, data.PrevPage, data.Category, data.Language)
	}
	
	fmt.Fprintf(w, `<span class="pagination-info">Page %d of %d</span>`, data.CurrentPage, data.TotalPages)
	
	if data.HasNext {
		fmt.Fprintf(w, `
		<button class="pagination-btn" 
				hx-get="/news?page=%d&category=%s&language=%s"
				hx-target="#news-content" 
				hx-swap="innerHTML">
			Next
			<img src="/static/images/svg/arrow-right-circle.svg" alt="Next" />
		</button>`, data.NextPage, data.Category, data.Language)
	}
	
	fmt.Fprintf(w, `</div>`)
}

// RefreshNewsHandler manually triggers news fetching
func RefreshNewsHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get source ID from URL if specified
	vars := mux.Vars(r)
	sourceIDStr := vars["sourceId"]

	fetchService := news.NewFetchService(db.DB)

	if sourceIDStr != "" {
		// Fetch from specific source
		sourceID, err := uuid.Parse(sourceIDStr)
		if err != nil {
			http.Error(w, "Invalid source ID", http.StatusBadRequest)
			return
		}

		err = fetchService.FetchFromSourceManually(sourceID)
		if err != nil {
			http.Error(w, "Failed to fetch news", http.StatusInternalServerError)
			fmt.Printf("Error fetching from source %s: %v\n", sourceID, err)
			return
		}

		w.Header().Set("X-Toast-Message", "News refreshed successfully!")
		w.Header().Set("X-Toast-Type", "success")
	} else {
		// Trigger background fetch for all due sources
		go fetchService.StartBackgroundFetching()
		
		w.Header().Set("X-Toast-Message", "News refresh started in background!")
		w.Header().Set("X-Toast-Type", "info")
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		// Return just a success message for HTMX
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<div class="toast-message success">News refresh completed!</div>`)
		return
	}
	
	// Return updated articles list for non-HTMX requests
	http.Redirect(w, r, "/news", http.StatusSeeOther)
}

// NewsSourcesHandler shows news sources management
func NewsSourcesHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	user, err := db.GetUser(db.DB, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sources, err := db.GetNewsSourcesWithUserPrefs(db.DB, userID)
	if err != nil {
		http.Error(w, "Failed to load news sources", http.StatusInternalServerError)
		fmt.Printf("Error loading news sources: %v\n", err)
		return
	}

	data := struct {
		User    db.User
		Sources []db.NewsSourceWithUserPref
	}{
		User:    user,
		Sources: sources,
	}

	tmpl, err := template.ParseFiles("templates/base.html", "templates/news_sources.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		fmt.Printf("Template parsing error: %v\n", err)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		fmt.Printf("Template execution error: %v\n", err)
	}
}


// ToggleUserNewsPreferenceHandler toggles user's preference for a news source
func ToggleUserNewsPreferenceHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	// Get source ID from URL
	vars := mux.Vars(r)
	sourceIDStr := vars["sourceId"]
	sourceID, err := uuid.Parse(sourceIDStr)
	if err != nil {
		http.Error(w, "Invalid source ID", http.StatusBadRequest)
		return
	}

	// Get current user preference
	userPrefs, err := db.GetUserNewsPreferences(db.DB, userID)
	if err != nil {
		http.Error(w, "Failed to get user preferences", http.StatusInternalServerError)
		fmt.Printf("Error getting user preferences: %v\n", err)
		return
	}

	// Determine current state (default to enabled if no preference exists)
	currentState, exists := userPrefs[sourceID]
	if !exists {
		currentState = true // Default to enabled
	}

	// Toggle the preference
	newState := !currentState
	err = db.SetUserNewsPreference(db.DB, userID, sourceID, newState)
	if err != nil {
		http.Error(w, "Failed to update user preference", http.StatusInternalServerError)
		fmt.Printf("Error updating user preference: %v\n", err)
		return
	}

	// Return the updated user preference button HTML
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if newState {
		fmt.Fprintf(w, `<button class="user-pref-btn enabled" 
			hx-post="/news/sources/%s/user-toggle" 
			hx-trigger="click"
			hx-swap="outerHTML"
			title="Toggle in Your Feed">
			<img src="/static/images/svg/view.svg" alt="Shown" />
			Show in Feed
		</button>`, sourceID)
	} else {
		fmt.Fprintf(w, `<button class="user-pref-btn disabled" 
			hx-post="/news/sources/%s/user-toggle" 
			hx-trigger="click"
			hx-swap="outerHTML"
			title="Toggle in Your Feed">
			<img src="/static/images/svg/x.svg" alt="Hidden" />
			Hide from Feed
		</button>`, sourceID)
	}
}

// FullArticleHandler displays the full article view
func FullArticleHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userIDValue := r.Context().Value("userID")
	if userIDValue == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := db.GetUser(db.DB, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get article ID from URL
	vars := mux.Vars(r)
	articleIDStr := vars["id"]
	articleID, err := uuid.Parse(articleIDStr)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	// Get article from database
	article, err := db.GetNewsArticleByID(db.DB, articleID)
	if err != nil {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}

	// Convert to template article with processed content
	fullContent := stringPtrToString(article.FullContent)
	content := stringPtrToString(article.Content)
	
	// Process line breaks for display
	if fullContent != "" {
		fullContent = strings.ReplaceAll(fullContent, "\r\n", "\n")
		fullContent = strings.ReplaceAll(fullContent, "\r", "\n")
	}
	if content != "" {
		content = strings.ReplaceAll(content, "\r\n", "\n")
		content = strings.ReplaceAll(content, "\r", "\n")
	}
	
	templateArticle := TemplateArticle{
		ID:           article.ID.String(),
		Title:        article.Title,
		Content:      content,
		Summary:      stringPtrToString(article.Summary),
		FullContent:  fullContent,
		OriginalURL:  article.OriginalURL,
		ImageURL:     stringPtrToString(article.ImageURL),
		ThumbnailURL: stringPtrToString(article.ThumbnailURL),
		AuthorName:   stringPtrToString(article.AuthorName),
		Keywords:     stringPtrToString(article.Keywords),
		PublishedAt:  timePtrToString(article.PublishedAt),
		ModifiedAt:   timePtrToString(article.DateModified),
		Language:     article.Language,
		Category:     article.Category,
		SourceName:   article.SourceName,
	}

	data := struct {
		User    db.User
		Article TemplateArticle
	}{
		User:    user,
		Article: templateArticle,
	}

	// Render full article template
	tmpl, err := template.ParseFiles("templates/base.html", "templates/full_article.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		fmt.Printf("Template parsing error: %v\n", err)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		fmt.Printf("Template execution error: %v\n", err)
	}
}