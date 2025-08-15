package news

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// RSS Feed Structures
type RSSFeed struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
	Content     string `xml:"content"`
}

// ParsedArticle represents a news article after RSS parsing
type ParsedArticle struct {
	Title       string
	Content     string
	URL         string
	PublishedAt time.Time
	ImageURL    string
}

// RSSParser handles RSS feed parsing
type RSSParser struct {
	client *http.Client
}

// NewRSSParser creates a new RSS parser instance
func NewRSSParser() *RSSParser {
	return &RSSParser{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ParseRSSFeed fetches and parses an RSS feed from the given URL
func (p *RSSParser) ParseRSSFeed(url string) ([]ParsedArticle, error) {
	// Fetch RSS feed
	resp, err := p.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSS feed from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RSS feed returned status %d for URL %s", resp.StatusCode, url)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read RSS response: %w", err)
	}

	// Parse XML
	var feed RSSFeed
	err = xml.Unmarshal(body, &feed)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSS XML: %w", err)
	}

	// Convert RSS items to parsed articles
	var articles []ParsedArticle
	for _, item := range feed.Channel.Items {
		article, err := p.convertRSSItem(item)
		if err != nil {
			// Log error but continue with other articles
			fmt.Printf("Error converting RSS item: %v\n", err)
			continue
		}
		articles = append(articles, article)
	}

	return articles, nil
}

// convertRSSItem converts an RSS item to a ParsedArticle
func (p *RSSParser) convertRSSItem(item Item) (ParsedArticle, error) {
	// Parse publication date
	pubDate, err := p.parseRSSDate(item.PubDate)
	if err != nil {
		// Use current time in Kuwait timezone if parsing fails
		kuwaitTZ, tzErr := time.LoadLocation("Asia/Kuwait")
		if tzErr != nil {
			kuwaitTZ = time.FixedZone("AST", 3*3600) // Arabia Standard Time (UTC+3)
		}
		pubDate = time.Now().In(kuwaitTZ)
	}

	// Clean content - prefer content over description
	content := item.Content
	if content == "" {
		content = item.Description
	}
	content = p.cleanHTMLContent(content)

	// Extract image URL from content if available
	imageURL := p.extractImageURL(content)

	return ParsedArticle{
		Title:       strings.TrimSpace(item.Title),
		Content:     content,
		URL:         strings.TrimSpace(item.Link),
		PublishedAt: pubDate,
		ImageURL:    imageURL,
	}, nil
}

// parseRSSDate handles various RSS date formats and converts to Kuwait time
func (p *RSSParser) parseRSSDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}

	// Kuwait timezone
	kuwaitTZ, err := time.LoadLocation("Asia/Kuwait")
	if err != nil {
		// Fallback to UTC+3 if Kuwait timezone is not available
		kuwaitTZ = time.FixedZone("AST", 3*3600) // Arabia Standard Time (UTC+3)
	}

	// Common RSS date formats
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		"Mon, 02 Jan 2006 15:04:05 -0700",
		"Mon, 02 Jan 2006 15:04:05 MST",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			// If the time already has timezone info and it's +0300 (Kuwait), keep it as is
			if t.Location().String() != "UTC" {
				_, offset := t.Zone()
				if offset == 3*3600 { // Already Kuwait time (+3 hours)
					return t, nil
				}
			}
			// Otherwise convert to Kuwait time
			return t.In(kuwaitTZ), nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// cleanHTMLContent removes HTML tags and cleans up content
func (p *RSSParser) cleanHTMLContent(content string) string {
	if content == "" {
		return ""
	}

	// Remove HTML tags (basic cleanup)
	content = strings.ReplaceAll(content, "<![CDATA[", "")
	content = strings.ReplaceAll(content, "]]>", "")
	
	// Remove common HTML tags while preserving content
	htmlTags := []string{
		"<p>", "</p>", "<br>", "<br/>", "<div>", "</div>",
		"<span>", "</span>", "<strong>", "</strong>", "<b>", "</b>",
		"<em>", "</em>", "<i>", "</i>", "<u>", "</u>",
	}
	
	for _, tag := range htmlTags {
		content = strings.ReplaceAll(content, tag, " ")
	}

	// Clean up multiple spaces and line breaks
	content = strings.ReplaceAll(content, "\n", " ")
	content = strings.ReplaceAll(content, "\r", " ")
	
	// Replace multiple spaces with single space
	for strings.Contains(content, "  ") {
		content = strings.ReplaceAll(content, "  ", " ")
	}

	// Remove any invalid UTF-8 characters that might cause issues
	content = strings.ToValidUTF8(content, "")

	return strings.TrimSpace(content)
}

// extractImageURL attempts to extract an image URL from content
func (p *RSSParser) extractImageURL(content string) string {
	// Look for img tags in the content
	if strings.Contains(content, "<img") {
		// Simple regex alternative using string operations
		imgStart := strings.Index(content, "<img")
		if imgStart != -1 {
			imgTag := content[imgStart:]
			imgEnd := strings.Index(imgTag, ">")
			if imgEnd != -1 {
				imgTag = imgTag[:imgEnd+1]
				
				// Extract src attribute
				srcStart := strings.Index(imgTag, "src=\"")
				if srcStart != -1 {
					srcStart += 5 // Skip 'src="'
					srcEnd := strings.Index(imgTag[srcStart:], "\"")
					if srcEnd != -1 {
						return imgTag[srcStart : srcStart+srcEnd]
					}
				}
			}
		}
	}
	return ""
}