package news

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// StructuredData represents JSON-LD structured data from news articles
type StructuredData struct {
	Context         string    `json:"@context"`
	Type            string    `json:"@type"`
	DateCreated     string    `json:"dateCreated"`
	DatePublished   string    `json:"datePublished"`
	DateModified    string    `json:"dateModified"`
	URL             string    `json:"url"`
	Headline        string    `json:"headline"`
	Description     string    `json:"description"`
	Keywords        string    `json:"keywords"`
	InLanguage      string    `json:"inLanguage"`
	ThumbnailURL    string    `json:"thumbnailUrl"`
	Image           ImageData `json:"image"`
	ArticleBody     string    `json:"articleBody"`
	Author          AuthorData `json:"author"`
	Publisher       PublisherData `json:"publisher"`
}

type ImageData struct {
	Type string `json:"@type"`
	URL  string `json:"url"`
}

type AuthorData struct {
	Type string `json:"@type"`
	Name string `json:"name"`
}

type PublisherData struct {
	Type string `json:"@type"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// EnhancedArticle represents a full article with all extracted data
type EnhancedArticle struct {
	Title         string
	FullContent   string
	Summary       string
	URL           string
	ThumbnailURL  string
	ImageURL      string
	AuthorName    string
	Keywords      string
	PublishedAt   time.Time
	ModifiedAt    time.Time
}

// ArticleScraper handles full article content extraction
type ArticleScraper struct {
	client *http.Client
}

// NewArticleScraper creates a new article scraper instance
func NewArticleScraper() *ArticleScraper {
	return &ArticleScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ScrapeFullArticle fetches and extracts full article content from URL
func (s *ArticleScraper) ScrapeFullArticle(url string) (*EnhancedArticle, error) {
	// Fetch article page
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch article page from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("article page returned status %d for URL %s", resp.StatusCode, url)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read article response: %w", err)
	}

	htmlContent := string(body)

	// Extract JSON-LD structured data
	structuredData, err := s.extractJSONLD(htmlContent)
	if err != nil {
		return nil, fmt.Errorf("failed to extract structured data: %w", err)
	}

	// Convert to EnhancedArticle
	return s.convertToEnhancedArticle(structuredData, url)
}

// extractJSONLD extracts JSON-LD structured data from HTML
func (s *ArticleScraper) extractJSONLD(htmlContent string) (*StructuredData, error) {
	// Find all JSON-LD script tags
	re := regexp.MustCompile(`<script\s+type="application/ld\+json"[^>]*>\s*(\{[\s\S]*?\})\s*</script>`)
	matches := re.FindAllStringSubmatch(htmlContent, -1)
	
	if len(matches) == 0 {
		return nil, fmt.Errorf("no JSON-LD scripts found in HTML")
	}

	// Look for NewsArticle in any of the JSON-LD scripts
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		
		jsonData := match[1]
		jsonData = strings.TrimSpace(jsonData)
		
		// Check if this script contains NewsArticle
		if !strings.Contains(jsonData, `"@type"`) || !strings.Contains(jsonData, `"NewsArticle"`) {
			continue
		}
		
		// Clean up JSON data
		jsonData = s.cleanJSONData(jsonData)
		
		// Ensure valid UTF-8
		jsonData = strings.ToValidUTF8(jsonData, "")

		var structuredData StructuredData
		err := json.Unmarshal([]byte(jsonData), &structuredData)
		if err != nil {
			fmt.Printf("JSON parse error for data: %s\nError: %v\n", jsonData[:min(200, len(jsonData))], err)
			continue
		}

		// Check if this is actually a NewsArticle
		if structuredData.Type == "NewsArticle" {
			return &structuredData, nil
		}
	}
	
	return nil, fmt.Errorf("no JSON-LD NewsArticle found in %d scripts", len(matches))
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// cleanJSONData cleans and normalizes JSON-LD data
func (s *ArticleScraper) cleanJSONData(jsonData string) string {
	// Ensure valid UTF-8 first
	jsonData = strings.ToValidUTF8(jsonData, "")
	
	// Remove any trailing commas before closing braces/brackets
	jsonData = regexp.MustCompile(`,\s*}`).ReplaceAllString(jsonData, "}")
	jsonData = regexp.MustCompile(`,\s*]`).ReplaceAllString(jsonData, "]")
	
	// Remove carriage returns and normalize line breaks
	jsonData = strings.ReplaceAll(jsonData, "\r\n", "\n")
	jsonData = strings.ReplaceAll(jsonData, "\r", "\n")
	
	// Remove line breaks within string values that can break JSON parsing
	re := regexp.MustCompile(`"([^"]*)\n([^"]*)"`)
	for re.MatchString(jsonData) {
		jsonData = re.ReplaceAllString(jsonData, `"$1 $2"`)
	}
	
	// Handle any escaped characters
	jsonData = strings.ReplaceAll(jsonData, `\"`, `"`)
	
	// Remove any remaining invalid UTF-8 sequences
	jsonData = strings.ToValidUTF8(jsonData, "")
	
	return jsonData
}

// convertToEnhancedArticle converts StructuredData to EnhancedArticle
func (s *ArticleScraper) convertToEnhancedArticle(data *StructuredData, url string) (*EnhancedArticle, error) {
	// Parse dates with Kuwait timezone
	kuwaitTZ, err := time.LoadLocation("Asia/Kuwait")
	if err != nil {
		kuwaitTZ = time.FixedZone("AST", 3*3600) // Arabia Standard Time (UTC+3)
	}

	publishedAt := time.Now().In(kuwaitTZ)
	if data.DatePublished != "" {
		if t, err := time.Parse(time.RFC3339, data.DatePublished); err == nil {
			publishedAt = t.In(kuwaitTZ)
		}
	}

	modifiedAt := publishedAt
	if data.DateModified != "" {
		if t, err := time.Parse(time.RFC3339, data.DateModified); err == nil {
			modifiedAt = t.In(kuwaitTZ)
		}
	}

	// Determine best image URL (prefer main image over thumbnail)
	imageURL := data.Image.URL
	if imageURL == "" {
		imageURL = data.ThumbnailURL
	}

	// Clean and prepare content
	fullContent := strings.TrimSpace(data.ArticleBody)
	summary := data.Description
	if summary == "" && len(fullContent) > 200 {
		summary = fullContent[:200] + "..."
	}

	return &EnhancedArticle{
		Title:        strings.TrimSpace(data.Headline),
		FullContent:  fullContent,
		Summary:      summary,
		URL:          url,
		ThumbnailURL: data.ThumbnailURL,
		ImageURL:     imageURL,
		AuthorName:   data.Author.Name,
		Keywords:     data.Keywords,
		PublishedAt:  publishedAt,
		ModifiedAt:   modifiedAt,
	}, nil
}

// IsKuwaitNewsSource checks if URL is from supported Kuwait news sources
func (s *ArticleScraper) IsKuwaitNewsSource(url string) bool {
	supportedSources := []string{
		"aljarida.com",
		"alraimedia.com",
		"alrai.com",
	}
	
	for _, source := range supportedSources {
		if strings.Contains(url, source) {
			return true
		}
	}
	
	return false
}