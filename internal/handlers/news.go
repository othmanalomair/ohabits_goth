package handlers

import (
	"fmt"
	"html"
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
	Content      template.HTML
	Summary      string
	FullContent  template.HTML
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
	// Reddit-specific fields
	PostFlair    string // [D], [R], [P], [N], etc.
	RedditUser   string // /u/username format
	CommentCount string // for Reddit comment link
	CommentsURL  string // URL to Reddit comments
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
	SourceCount  int
	HasNext      bool
	HasPrev      bool
	NextPage     int
	PrevPage     int
	Sources      []db.NewsSourceWithUserPref
}

// convertToTemplateArticles converts database articles to template-friendly articles
func convertToTemplateArticles(articles []db.NewsArticle) []TemplateArticle {
	templateArticles := make([]TemplateArticle, len(articles))
	for i, article := range articles {
		// Process content based on category
		content := stringPtrToString(article.Content)
		fullContent := stringPtrToString(article.FullContent)
		
		// Extract Reddit-specific fields
		var postFlair, redditUser, commentCount, commentsURL string
		var extractedImageURL string
		
		if article.Category == "reddit" {
			// Extract Reddit metadata before processing content
			originalContent := stringPtrToString(article.Content)
			var metadata map[string]string
			
			// Extract submission metadata if it exists
			if strings.Contains(originalContent, "submitted by") {
				submittedIndex := strings.Index(originalContent, " submitted by ")
				if submittedIndex == -1 {
					submittedIndex = strings.Index(originalContent, "submitted by")
				}
				if submittedIndex != -1 {
					submissionInfo := originalContent[submittedIndex:]
					metadata = extractRedditSubmissionMetadata(submissionInfo)
				}
			}
			
			// Extract image from content if no main image exists
			hasMainImage := stringPtrToString(article.ImageURL) != ""
			
			if !hasMainImage {
				extractedImageURL = extractFirstImageFromRedditContent(originalContent)
				if extractedImageURL != "" {
					hasMainImage = true
				}
			}
			
			// For news list view, keep Reddit content simple - don't format with HTML
			// Only clean up basic entities but don't add complex HTML structures
			if fullContent != "" {
				fullContent = cleanRedditContentForList(fullContent)
			}
			if content != "" {
				content = cleanRedditContentForList(content)
			}
			
			// Extract Reddit-specific information
			postFlair = extractRedditFlair(article.Title)
			redditUser = extractRedditUser(stringPtrToString(article.AuthorName))
			
			// If no user in author field, try to extract from content or metadata
			if redditUser == "" {
				redditUser = extractRedditUserFromContent(content)
			}
			if redditUser == "" && metadata != nil && metadata["user"] != "" {
				redditUser = metadata["user"]
			}
			
			// Extract comment info
			commentCount = extractRedditCommentCount(content)
			if commentCount == "" && strings.Contains(article.OriginalURL, "reddit.com") {
				commentCount = "Discussion"
			}
			
			// Extract comments URL from metadata
			if metadata != nil && metadata["comments_url"] != "" {
				commentsURL = metadata["comments_url"]
			} else {
				commentsURL = article.OriginalURL
			}
		} else if article.Category == "hackernews" {
			// Format Hacker News content
			if fullContent != "" {
				fullContent = formatHackerNewsContent(fullContent)
			}
			if content != "" {
				content = formatHackerNewsContent(content)
			}
		}
		
		// Determine final image URL (use extracted image if available, otherwise original)
		finalImageURL := stringPtrToString(article.ImageURL)
		if extractedImageURL != "" {
			finalImageURL = extractedImageURL
		}
		
		// Decode HTML entities in the final image URL
		if finalImageURL != "" {
			finalImageURL = html.UnescapeString(finalImageURL)
		}
		
		templateArticles[i] = TemplateArticle{
			ID:           article.ID.String(),
			Title:        article.Title,
			Content:      template.HTML(content),
			Summary:      stringPtrToString(article.Summary),
			FullContent:  template.HTML(fullContent),
			OriginalURL:  article.OriginalURL,
			ImageURL:     finalImageURL,
			ThumbnailURL: stringPtrToString(article.ThumbnailURL),
			AuthorName:   stringPtrToString(article.AuthorName),
			Keywords:     stringPtrToString(article.Keywords),
			PublishedAt:  timePtrToString(article.PublishedAt),
			ModifiedAt:   timePtrToString(article.DateModified),
			Language:     article.Language,
			Category:     article.Category,
			SourceName:   article.SourceName,
			// Reddit-specific fields
			PostFlair:    postFlair,
			RedditUser:   redditUser,
			CommentCount: commentCount,
			CommentsURL:  commentsURL,
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

// formatHackerNewsContent formats HTML content specifically for Hacker News articles
func formatHackerNewsContent(content string) string {
	// Add specific classes for better styling
	formatted := strings.ReplaceAll(content, "<p>", "<p class=\"hn-paragraph\">")
	
	// Replace horizontal rules with styled separators
	formatted = strings.ReplaceAll(formatted, " <hr> ", "<hr class=\"hn-separator\" />")
	formatted = strings.ReplaceAll(formatted, "<hr>", "<hr class=\"hn-separator\" />")
	
	// Style links to break long URLs properly and open in new tab
	formatted = strings.ReplaceAll(formatted, "<a href=", "<a class=\"hn-link\" target=\"_blank\" rel=\"noopener noreferrer\" href=")
	
	// If there are no paragraph tags, try to add some structure by splitting on common patterns
	if !strings.Contains(formatted, "<p") {
		// Split on double line breaks if they exist
		if strings.Contains(formatted, "\n\n") {
			parts := strings.Split(formatted, "\n\n")
			var formattedParts []string
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part != "" {
					formattedParts = append(formattedParts, "<p class=\"hn-paragraph\">"+part+"</p>")
				}
			}
			formatted = strings.Join(formattedParts, "\n")
		} else {
			// If no double line breaks, look for sentence endings followed by capital letters
			// This is a simple heuristic that works reasonably well for HN posts
			formatted = strings.ReplaceAll(formatted, ". ", ".</p><p class=\"hn-paragraph\">")
			formatted = "<p class=\"hn-paragraph\">" + formatted + "</p>"
			
			// Clean up any malformed paragraphs
			formatted = strings.ReplaceAll(formatted, "<p class=\"hn-paragraph\"></p>", "")
			formatted = strings.ReplaceAll(formatted, "<p class=\"hn-paragraph\"><hr", "<hr")
		}
	}
	
	return formatted
}

// formatRedditContent formats HTML content specifically for Reddit posts
func formatRedditContent(content string) string {
	// Remove Reddit's SC_OFF/SC_ON comments
	formatted := strings.ReplaceAll(content, "<!-- SC_OFF -->", "")
	formatted = strings.ReplaceAll(formatted, "<!-- SC_ON -->", "")
	
	// Convert HTML table structure (used by Reddit for posts with images) to cleaner format
	if strings.Contains(formatted, "<table>") {
		// Extract image from table structure
		formatted = extractRedditTableContent(formatted)
	}
	
	// Style Reddit markdown divs
	formatted = strings.ReplaceAll(formatted, "<div class=\"md\">", "<div class=\"reddit-content\">")
	
	// Style paragraphs for Reddit
	formatted = strings.ReplaceAll(formatted, "<p>", "<p class=\"reddit-paragraph\">")
	
	// Style links to open in new tab
	formatted = strings.ReplaceAll(formatted, "<a href=", "<a class=\"reddit-link\" target=\"_blank\" rel=\"noopener noreferrer\" href=")
	
	// Style strong/bold text
	formatted = strings.ReplaceAll(formatted, "<strong>", "<strong class=\"reddit-bold\">")
	
	// Clean up Reddit submission footer
	if strings.Contains(formatted, "submitted by") {
		// Extract and clean the submission info
		formatted = cleanRedditSubmissionInfo(formatted)
	}
	
	return formatted
}

// formatRedditContentWithImageCheck formats HTML content for Reddit posts with proper formatting and image handling
func formatRedditContentWithImageCheck(content string, hasImageAlready bool) string {
	// First decode HTML entities if they exist
	formatted := content
	if strings.Contains(formatted, "&lt;") || strings.Contains(formatted, "&gt;") || strings.Contains(formatted, "&quot;") {
		formatted = strings.ReplaceAll(formatted, "&lt;", "<")
		formatted = strings.ReplaceAll(formatted, "&gt;", ">")
		formatted = strings.ReplaceAll(formatted, "&quot;", "\"")
		formatted = strings.ReplaceAll(formatted, "&amp;", "&")
		formatted = strings.ReplaceAll(formatted, "&#32;", " ")
	}
	
	// Remove Reddit's SC_OFF/SC_ON comments
	formatted = strings.ReplaceAll(formatted, "<!-- SC_OFF -->", "")
	formatted = strings.ReplaceAll(formatted, "<!-- SC_ON -->", "")
	
	// Convert HTML table structure to cleaner format
	if strings.Contains(formatted, "<table>") {
		formatted = extractAndFormatRedditTableContent(formatted, hasImageAlready)
	}
	
	// Process Reddit submission footer BEFORE image processing to extract images cleanly
	if strings.Contains(formatted, "submitted by") {
		formatted = formatRedditSubmissionInfo(formatted)
	}
	
	// Handle image links - if we have a main image, remove all image links; otherwise convert them
	if hasImageAlready {
		// Remove all image links since we have a main image displayed
		formatted = removeAllImageLinksFromContent(formatted)
	} else {
		// Convert image links to inline images
		formatted = convertImageLinksToImages(formatted)
	}
	
	// Add video preview support for Reddit posts
	formatted = addVideoPreviewSupport(formatted)
	
	// Style Reddit markdown divs
	formatted = strings.ReplaceAll(formatted, "<div class=\"md\">", "<div class=\"reddit-content\">")
	
	// Convert paragraph tags to line breaks for better Reddit formatting
	formatted = strings.ReplaceAll(formatted, "<p>", "")
	formatted = strings.ReplaceAll(formatted, "</p>", "<br><br>")
	
	// Add proper line breaks - Reddit uses "; -" as bullet points
	formatted = strings.ReplaceAll(formatted, "; -", ";<br><br>•")
	formatted = strings.ReplaceAll(formatted, "- ", "<br>• ")
	
	// Add line breaks after sentences for better readability
	formatted = strings.ReplaceAll(formatted, ". ", ".<br><br>")
	formatted = strings.ReplaceAll(formatted, "? ", "?<br><br>")
	formatted = strings.ReplaceAll(formatted, "! ", "!<br><br>")
	
	// Convert double spaces to line breaks for better readability  
	formatted = strings.ReplaceAll(formatted, "  ", "<br>")
	
	// Style links to open in new tab
	formatted = strings.ReplaceAll(formatted, "<a href=", "<a class=\"reddit-link\" target=\"_blank\" rel=\"noopener noreferrer\" href=")
	
	return formatted
}

// cleanRedditContentForList cleans Reddit content for news list display without complex HTML
func cleanRedditContentForList(content string) string {
	// Handle table structure first - extract text content from table
	if strings.Contains(content, "<table>") {
		content = extractTextFromTable(content)
	}
	
	// Remove HTML comments
	cleaned := strings.ReplaceAll(content, "<!-- SC_OFF -->", "")
	cleaned = strings.ReplaceAll(cleaned, "<!-- SC_ON -->", "")
	
	// Remove submitted by sections entirely for news list
	submittedIndex := strings.Index(cleaned, " submitted by ")
	if submittedIndex == -1 {
		submittedIndex = strings.Index(cleaned, "submitted by")
	}
	if submittedIndex != -1 {
		cleaned = cleaned[:submittedIndex]
	}
	
	// Remove image links for news list view
	cleaned = removeAllImageLinksFromContent(cleaned)
	
	// Clean up basic HTML entities
	cleaned = html.UnescapeString(cleaned)
	
	// Remove ALL HTML tags for news list - be more aggressive
	// Remove table remnants
	cleaned = strings.ReplaceAll(cleaned, "<table>", "")
	cleaned = strings.ReplaceAll(cleaned, "</table>", "")
	cleaned = strings.ReplaceAll(cleaned, "<tr>", "")
	cleaned = strings.ReplaceAll(cleaned, "</tr>", "")
	cleaned = strings.ReplaceAll(cleaned, "<td>", "")
	cleaned = strings.ReplaceAll(cleaned, "</td>", " ")
	cleaned = strings.ReplaceAll(cleaned, "<th>", "")
	cleaned = strings.ReplaceAll(cleaned, "</th>", " ")
	
	// Remove common HTML tags
	cleaned = strings.ReplaceAll(cleaned, "<div class=\"md\">", "")
	cleaned = strings.ReplaceAll(cleaned, "<div>", "")
	cleaned = strings.ReplaceAll(cleaned, "</div>", "")
	cleaned = strings.ReplaceAll(cleaned, "<p>", "")
	cleaned = strings.ReplaceAll(cleaned, "</p>", " ")
	cleaned = strings.ReplaceAll(cleaned, "<br>", " ")
	cleaned = strings.ReplaceAll(cleaned, "<br/>", " ")
	
	// Remove any remaining anchor tags but keep the text
	cleaned = strings.ReplaceAll(cleaned, "</a>", "")
	// Remove opening anchor tags with href
	for strings.Contains(cleaned, "<a href=") {
		start := strings.Index(cleaned, "<a href=")
		if start == -1 {
			break
		}
		end := strings.Index(cleaned[start:], ">")
		if end == -1 {
			break
		}
		cleaned = cleaned[:start] + cleaned[start+end+1:]
	}
	
	// Clean up multiple spaces
	for strings.Contains(cleaned, "  ") {
		cleaned = strings.ReplaceAll(cleaned, "  ", " ")
	}
	
	return strings.TrimSpace(cleaned)
}

// extractAndFormatRedditTableContent handles table content with proper image and text formatting
func extractAndFormatRedditTableContent(content string, hasImageAlready bool) string {
	// Extract text content from table
	textContent := extractTextFromTable(content)
	
	if hasImageAlready {
		// Just return cleaned text if we already have the main image
		return textContent
	}
	
	// Extract and format images from table if no main image exists
	if strings.Contains(content, "<img src=") {
		start := strings.Index(content, "<img src=\"")
		if start != -1 {
			start += 10
			end := strings.Index(content[start:], "\"")
			if end != -1 {
				imageURL := content[start : start+end]
				return fmt.Sprintf(`<div class="reddit-content-image">
					<img src="%s" class="reddit-inline-image" alt="Reddit content image" />
				</div>
				<div class="reddit-text-content">%s</div>`, imageURL, textContent)
			}
		}
	}
	
	return textContent
}

// convertImageLinksToImages converts image URLs in content to actual img tags
func convertImageLinksToImages(content string) string {
	// Convert image URLs that are just text links to actual images
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	imageDomains := []string{"i.redd.it", "preview.redd.it", "imgur.com", "i.imgur.com"}
	
	// Pattern: <a href="IMAGE_URL">[link]</a>
	// Replace with actual images
	for _, domain := range imageDomains {
		for _, ext := range imageExtensions {
			// Look for Reddit image links
			pattern := `<a href="https://` + domain + `/[^"]*` + ext + `">\[link\]</a>`
			content = replaceImageLinksWithImages(content, pattern, domain, ext)
		}
	}
	
	// Also handle direct image links that aren't wrapped with [link]
	for _, ext := range imageExtensions {
		if strings.Contains(content, ext) {
			content = highlightImageLinks(content, ext)
		}
	}
	
	return content
}

// replaceImageLinksWithImages replaces specific image link patterns with actual img tags
func replaceImageLinksWithImages(content, pattern, domain, ext string) string {
	// Use a simple string search approach since we don't have regex
	searchTerm := `<a href="https://` + domain
	endTerm := ext + `">[link]</a>`
	
	searchPos := 0
	for {
		start := strings.Index(content[searchPos:], searchTerm)
		if start == -1 {
			break
		}
		start += searchPos
		
		// Find the end of this link
		end := strings.Index(content[start:], endTerm)
		if end == -1 {
			searchPos = start + len(searchTerm)
			continue
		}
		
		// Extract the full URL
		urlStart := start + len(`<a href="`)
		urlEnd := strings.Index(content[urlStart:], `">`)
		if urlEnd == -1 {
			searchPos = start + len(searchTerm)
			continue
		}
		
		imageURL := content[urlStart : urlStart+urlEnd]
		
		// Replace with image tag that will be prominently displayed
		imageHTML := fmt.Sprintf(`<div class="reddit-content-image">
			<img src="%s" class="reddit-inline-image" alt="Reddit image" loading="lazy" />
		</div>`, imageURL)
		
		fullLinkEnd := start + end + len(endTerm)
		content = content[:start] + imageHTML + content[fullLinkEnd:]
		
		// Update search position
		searchPos = start + len(imageHTML)
	}
	
	return content
}

// extractFirstImageFromRedditContent extracts the first image URL from Reddit content to use as main image
func extractFirstImageFromRedditContent(content string) string {
	// First decode HTML entities
	content = strings.ReplaceAll(content, "&#32;", " ")
	content = strings.ReplaceAll(content, "&amp;", "&")
	content = strings.ReplaceAll(content, "&lt;", "<")
	content = strings.ReplaceAll(content, "&gt;", ">")
	content = strings.ReplaceAll(content, "&quot;", "\"")
	
	imageDomains := []string{"i.redd.it", "preview.redd.it", "imgur.com", "i.imgur.com"}
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	
	for _, domain := range imageDomains {
		searchTerm := `<a href="https://` + domain
		
		searchPos := 0
		start := strings.Index(content[searchPos:], searchTerm)
		if start == -1 {
			continue
		}
		start += searchPos
		
		// Extract the full URL
		urlStart := start + len(`<a href="`)
		urlEnd := strings.Index(content[urlStart:], `">`)
		if urlEnd == -1 {
			continue
		}
		
		imageURL := content[urlStart : urlStart+urlEnd]
		
		// Check if it's actually an image
		for _, ext := range imageExtensions {
			if strings.Contains(imageURL, ext) {
				return imageURL
			}
		}
	}
	
	return ""
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// highlightImageLinks highlights remaining image links that weren't converted
func highlightImageLinks(content, ext string) string {
	// Style remaining image links specially
	content = strings.ReplaceAll(content, ext+`">`, ext+`" class="reddit-image-link">`)
	return content
}

// addVideoPreviewSupport adds video preview support for various platforms
func addVideoPreviewSupport(content string) string {
	var videoLinks []string
	
	// Extract video links and collect them
	content, videoLinks = extractAndEnhanceVideoLinks(content, videoLinks)
	
	// If we found video links, prepend them to the content for prominence
	if len(videoLinks) > 0 {
		videoSection := "<div class=\"reddit-video-section\">"
		for _, videoLink := range videoLinks {
			videoSection += videoLink
		}
		videoSection += "</div><br><br>"
		content = videoSection + content
	}
	
	return content
}

// extractAndEnhanceVideoLinks extracts and enhances video links, returning updated content and collected links
func extractAndEnhanceVideoLinks(content string, videoLinks []string) (string, []string) {
	// Handle YouTube links
	content, videoLinks = enhanceYouTubeLinksWithExtraction(content, videoLinks)
	
	// Handle PeerTube links
	content, videoLinks = enhancePeerTubeLinksWithExtraction(content, videoLinks)
	
	// Handle Vimeo links
	content, videoLinks = enhanceVimeoLinksWithExtraction(content, videoLinks)
	
	// Handle Reddit video links  
	content, videoLinks = enhanceRedditVideoLinksWithExtraction(content, videoLinks)
	
	return content, videoLinks
}

// enhanceYouTubeLinksWithExtraction extracts YouTube links for top placement
func enhanceYouTubeLinksWithExtraction(content string, videoLinks []string) (string, []string) {
	ytDomains := []string{"youtube.com", "youtu.be", "www.youtube.com"}
	
	for _, domain := range ytDomains {
		searchTerm := `<a href="https://` + domain
		searchPos := 0
		
		for {
			start := strings.Index(content[searchPos:], searchTerm)
			if start == -1 {
				break
			}
			start += searchPos
			
			// Find the end of the link
			end := strings.Index(content[start:], "</a>")
			if end == -1 {
				break
			}
			
			linkContent := content[start : start+end+4]
			
			// Create enhanced video link
			enhancedLink := strings.ReplaceAll(linkContent, `<a href=`, `<a class="reddit-video-link youtube-link" href=`)
			enhancedLink = strings.ReplaceAll(enhancedLink, `">`, `"><span class="video-icon">🎥</span>`)
			enhancedLink = strings.ReplaceAll(enhancedLink, `</a>`, ` <span class="video-platform">YouTube</span></a>`)
			
			// Add to video links collection
			videoLinks = append(videoLinks, enhancedLink)
			
			// Remove from original content
			content = content[:start] + content[start+end+4:]
			// Don't increment searchPos since we removed content
		}
	}
	
	return content, videoLinks
}

// enhanceYouTubeLinks adds preview info for YouTube links (legacy function)
func enhanceYouTubeLinks(content string) string {
	// Look for YouTube links and add preview styling
	ytDomains := []string{"youtube.com", "youtu.be", "www.youtube.com"}
	
	for _, domain := range ytDomains {
		searchTerm := `<a href="https://` + domain
		searchPos := 0
		
		for {
			start := strings.Index(content[searchPos:], searchTerm)
			if start == -1 {
				break
			}
			start += searchPos
			
			// Find the end of the link
			end := strings.Index(content[start:], "</a>")
			if end == -1 {
				break
			}
			
			linkContent := content[start : start+end+4]
			
			// Add video preview styling
			enhancedLink := strings.ReplaceAll(linkContent, `<a href=`, `<a class="reddit-video-link youtube-link" href=`)
			enhancedLink = strings.ReplaceAll(enhancedLink, `">`, `"><span class="video-icon">🎥</span>`)
			
			content = content[:start] + enhancedLink + content[start+end+4:]
			searchPos = start + len(enhancedLink)
		}
	}
	
	return content
}

// enhancePeerTubeLinksWithExtraction extracts PeerTube links for top placement
func enhancePeerTubeLinksWithExtraction(content string, videoLinks []string) (string, []string) {
	peerTubeDomains := []string{"peertube.wtf", "framatube.org", "video.antopie.org"}
	
	for _, domain := range peerTubeDomains {
		searchTerm := `<a href="https://` + domain
		searchPos := 0
		
		for {
			start := strings.Index(content[searchPos:], searchTerm)
			if start == -1 {
				break
			}
			start += searchPos
			
			// Find the end of the link
			end := strings.Index(content[start:], "</a>")
			if end == -1 {
				break
			}
			
			linkContent := content[start : start+end+4]
			
			// Create enhanced video link
			enhancedLink := strings.ReplaceAll(linkContent, `<a href=`, `<a class="reddit-video-link peertube-link" href=`)
			enhancedLink = strings.ReplaceAll(enhancedLink, `">`, `"><span class="video-icon">📹</span><span class="video-platform">PeerTube</span>`)
			
			// Add to video links collection
			videoLinks = append(videoLinks, enhancedLink)
			
			// Remove from original content
			content = content[:start] + content[start+end+4:]
			// Don't increment searchPos since we removed content
		}
	}
	
	return content, videoLinks
}

// enhancePeerTubeLinks adds preview info for PeerTube links (legacy function)
func enhancePeerTubeLinks(content string) string {
	// Look for PeerTube domains
	peerTubeDomains := []string{"peertube.wtf", "framatube.org", "video.antopie.org"}
	
	for _, domain := range peerTubeDomains {
		searchTerm := `<a href="https://` + domain
		searchPos := 0
		
		for {
			start := strings.Index(content[searchPos:], searchTerm)
			if start == -1 {
				break
			}
			start += searchPos
			
			// Find the end of the link
			end := strings.Index(content[start:], "</a>")
			if end == -1 {
				break
			}
			
			linkContent := content[start : start+end+4]
			
			// Add PeerTube preview styling
			enhancedLink := strings.ReplaceAll(linkContent, `<a href=`, `<a class="reddit-video-link peertube-link" href=`)
			enhancedLink = strings.ReplaceAll(enhancedLink, `">`, `"><span class="video-icon">📹</span><span class="video-platform">PeerTube</span>`)
			
			content = content[:start] + enhancedLink + content[start+end+4:]
			searchPos = start + len(enhancedLink)
		}
	}
	
	return content
}

// enhanceVimeoLinksWithExtraction extracts Vimeo links for top placement
func enhanceVimeoLinksWithExtraction(content string, videoLinks []string) (string, []string) {
	searchTerm := `<a href="https://vimeo.com`
	searchPos := 0
	
	for {
		start := strings.Index(content[searchPos:], searchTerm)
		if start == -1 {
			break
		}
		start += searchPos
		
		// Find the end of the link
		end := strings.Index(content[start:], "</a>")
		if end == -1 {
			break
		}
		
		linkContent := content[start : start+end+4]
		
		// Create enhanced video link
		enhancedLink := strings.ReplaceAll(linkContent, `<a href=`, `<a class="reddit-video-link vimeo-link" href=`)
		enhancedLink = strings.ReplaceAll(enhancedLink, `">`, `"><span class="video-icon">🎬</span><span class="video-platform">Vimeo</span>`)
		
		// Add to video links collection
		videoLinks = append(videoLinks, enhancedLink)
		
		// Remove from original content
		content = content[:start] + content[start+end+4:]
		// Don't increment searchPos since we removed content
	}
	
	return content, videoLinks
}

// enhanceRedditVideoLinksWithExtraction extracts Reddit video links for top placement
func enhanceRedditVideoLinksWithExtraction(content string, videoLinks []string) (string, []string) {
	redditVideoDomains := []string{"v.redd.it", "reddit.com/video"}
	
	for _, domain := range redditVideoDomains {
		searchTerm := `<a href="https://` + domain
		if domain == "reddit.com/video" {
			searchTerm = `<a href="https://www.reddit.com/video`
		}
		
		searchPos := 0
		
		for {
			start := strings.Index(content[searchPos:], searchTerm)
			if start == -1 {
				break
			}
			start += searchPos
			
			// Find the end of the link
			end := strings.Index(content[start:], "</a>")
			if end == -1 {
				break
			}
			
			linkContent := content[start : start+end+4]
			
			// Create enhanced video link
			enhancedLink := strings.ReplaceAll(linkContent, `<a href=`, `<a class="reddit-video-link reddit-video" href=`)
			enhancedLink = strings.ReplaceAll(enhancedLink, `">`, `"><span class="video-icon">🎞️</span><span class="video-platform">Reddit Video</span>`)
			
			// Add to video links collection
			videoLinks = append(videoLinks, enhancedLink)
			
			// Remove from original content
			content = content[:start] + content[start+end+4:]
			// Don't increment searchPos since we removed content
		}
	}
	
	return content, videoLinks
}

// enhanceVimeoLinks adds preview info for Vimeo links
func enhanceVimeoLinks(content string) string {
	searchTerm := `<a href="https://vimeo.com`
	searchPos := 0
	
	for {
		start := strings.Index(content[searchPos:], searchTerm)
		if start == -1 {
			break
		}
		start += searchPos
		
		// Find the end of the link
		end := strings.Index(content[start:], "</a>")
		if end == -1 {
			break
		}
		
		linkContent := content[start : start+end+4]
		
		// Add Vimeo preview styling
		enhancedLink := strings.ReplaceAll(linkContent, `<a href=`, `<a class="reddit-video-link vimeo-link" href=`)
		enhancedLink = strings.ReplaceAll(enhancedLink, `">`, `"><span class="video-icon">🎬</span><span class="video-platform">Vimeo</span>`)
		
		content = content[:start] + enhancedLink + content[start+end+4:]
		searchPos = start + len(enhancedLink)
	}
	
	return content
}

// enhanceRedditVideoLinks adds preview info for Reddit video links
func enhanceRedditVideoLinks(content string) string {
	redditVideoDomains := []string{"v.redd.it", "reddit.com/video"}
	
	for _, domain := range redditVideoDomains {
		searchTerm := `<a href="https://` + domain
		if domain == "reddit.com/video" {
			searchTerm = `<a href="https://www.reddit.com/video`
		}
		
		searchPos := 0
		
		for {
			start := strings.Index(content[searchPos:], searchTerm)
			if start == -1 {
				break
			}
			start += searchPos
			
			// Find the end of the link
			end := strings.Index(content[start:], "</a>")
			if end == -1 {
				break
			}
			
			linkContent := content[start : start+end+4]
			
			// Add Reddit video preview styling
			enhancedLink := strings.ReplaceAll(linkContent, `<a href=`, `<a class="reddit-video-link reddit-video" href=`)
			enhancedLink = strings.ReplaceAll(enhancedLink, `">`, `"><span class="video-icon">🎞️</span><span class="video-platform">Reddit Video</span>`)
			
			content = content[:start] + enhancedLink + content[start+end+4:]
			searchPos = start + len(enhancedLink)
		}
	}
	
	return content
}

// removeAllImageLinksFromContent removes all image links from Reddit content 
func removeAllImageLinksFromContent(content string) string {
	imageDomains := []string{"i.redd.it", "preview.redd.it", "imgur.com", "i.imgur.com"}
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	
	// Remove image links - handle both [link] pattern and empty links
	for _, domain := range imageDomains {
		for _, ext := range imageExtensions {
			// Pattern 1: <a href="https://domain/path.ext">[link]</a>
			searchTerm := `<a href="https://` + domain
			endTerm := ext + `">[link]</a>`
			
			searchPos := 0
			for {
				start := strings.Index(content[searchPos:], searchTerm)
				if start == -1 {
					break
				}
				start += searchPos
				
				// Find the end of this link
				end := strings.Index(content[start:], endTerm)
				if end == -1 {
					searchPos = start + len(searchTerm)
					continue
				}
				
				fullLinkEnd := start + end + len(endTerm)
				// Remove the entire link
				content = content[:start] + content[fullLinkEnd:]
				// Don't increment searchPos since we removed content
			}
			
			// Pattern 2: <a class="reddit-link" ... href="https://domain/path.ext"></a> (empty links)
			searchPos = 0
			for {
				linkStart := strings.Index(content[searchPos:], `<a`)
				if linkStart == -1 {
					break
				}
				linkStart += searchPos
				
				linkEnd := strings.Index(content[linkStart:], `</a>`)
				if linkEnd == -1 {
					break
				}
				
				fullLink := content[linkStart : linkStart+linkEnd+4]
				
				// Check if this link contains our image domain and extension
				if strings.Contains(fullLink, domain) && strings.Contains(fullLink, ext) {
					// Remove this link entirely
					content = content[:linkStart] + content[linkStart+linkEnd+4:]
					// Don't increment searchPos since we removed content
				} else {
					searchPos = linkStart + linkEnd + 4
				}
			}
		}
	}
	
	// Also remove any remaining image content divs
	content = strings.ReplaceAll(content, `<div class="reddit-content-image">`, "")
	
	// Remove img tags
	for strings.Contains(content, "<img") {
		start := strings.Index(content, "<img")
		if start == -1 {
			break
		}
		end := strings.Index(content[start:], ">")
		if end == -1 {
			break
		}
		content = content[:start] + content[start+end+1:]
	}
	
	// Clean up any leftover closing divs
	content = strings.ReplaceAll(content, `</div>`, "")
	
	return content
}

// removeDuplicateImagesButKeepOthers removes duplicate main images but keeps content images
func removeDuplicateImagesButKeepOthers(content string) string {
	// This is more selective - only remove images that match the main image
	// For now, we'll use the existing function but could be more sophisticated
	return removeImagesFromContent(content)
}

// extractRedditTableContent extracts content from Reddit's table structure
func extractRedditTableContent(content string) string {
	// Reddit uses tables for posts with images, extract the image and content separately
	if strings.Contains(content, "<img src=") {
		// Extract image URL
		start := strings.Index(content, "<img src=\"")
		if start != -1 {
			start += 10
			end := strings.Index(content[start:], "\"")
			if end != -1 {
				imageURL := content[start : start+end]
				// Create a cleaner structure
				textContent := extractTextFromTable(content)
				return fmt.Sprintf(`<div class="reddit-post-with-image">
					<img src="%s" class="reddit-post-image" alt="Post image" />
					<div class="reddit-post-text">%s</div>
				</div>`, imageURL, textContent)
			}
		}
	}
	// Fallback: just remove table tags
	formatted := strings.ReplaceAll(content, "<table>", "")
	formatted = strings.ReplaceAll(formatted, "</table>", "")
	formatted = strings.ReplaceAll(formatted, "<tr>", "")
	formatted = strings.ReplaceAll(formatted, "</tr>", "")
	formatted = strings.ReplaceAll(formatted, "<td>", "")
	formatted = strings.ReplaceAll(formatted, "</td>", "")
	return formatted
}

// extractTextFromTable extracts text content from Reddit table structure
func extractTextFromTable(content string) string {
	// Find the text content in the second td (after the image)
	start := strings.Index(content, "</td><td>")
	if start != -1 {
		start += 9
		end := strings.Index(content[start:], "</td>")
		if end != -1 {
			return content[start : start+end]
		}
	}
	return content
}

// cleanRedditSubmissionInfo cleans up Reddit submission footer
func cleanRedditSubmissionInfo(content string) string {
	// Remove the submission footer that contains "submitted by /u/username [link] [comments]"
	submittedIndex := strings.Index(content, " submitted by ")
	if submittedIndex != -1 {
		// Keep content before submission info
		content = content[:submittedIndex]
	}
	return content
}

// formatRedditSubmissionInfo formats Reddit submission footer to preserve useful links
func formatRedditSubmissionInfo(content string) string {
	// Try different patterns for submission info
	submittedIndex := strings.Index(content, " submitted by ")
	if submittedIndex == -1 {
		submittedIndex = strings.Index(content, "submitted by")
	}
	if submittedIndex == -1 {
		return content
	}
	
	// Extract main content before submission info
	mainContent := content[:submittedIndex]
	submissionInfo := content[submittedIndex:]
	
	// Format the submission info with proper styling
	submissionInfo = strings.ReplaceAll(submissionInfo, " submitted by ", "<br><br><div class=\"reddit-submission-info\">Posted by ")
	submissionInfo = strings.ReplaceAll(submissionInfo, "submitted by", "<br><br><div class=\"reddit-submission-info\">Posted by")
	
	// Clean up line breaks within submission info
	submissionInfo = strings.ReplaceAll(submissionInfo, "<br> <a", " <a")
	submissionInfo = strings.ReplaceAll(submissionInfo, "</a><br> <a", "</a> | <a")
	submissionInfo = strings.ReplaceAll(submissionInfo, "<br/>", " | ")
	
	// Format user link
	submissionInfo = strings.ReplaceAll(submissionInfo, " /u/", " /u/")
	
	// Remove [link] and format [comments]
	submissionInfo = strings.ReplaceAll(submissionInfo, "[link]</a>", "</a>")
	submissionInfo = strings.ReplaceAll(submissionInfo, "[comments]", "💬 Comments")
	
	// Close the submission info div
	if !strings.Contains(submissionInfo, "</div>") {
		submissionInfo += "</div>"
	}
	
	return mainContent + submissionInfo
}

// extractRedditSubmissionMetadata extracts Reddit submission metadata for top placement
func extractRedditSubmissionMetadata(submissionInfo string) map[string]string {
	metadata := make(map[string]string)
	
	// Extract username
	if strings.Contains(submissionInfo, "/u/") {
		userStart := strings.Index(submissionInfo, "/u/")
		if userStart != -1 {
			userPart := submissionInfo[userStart:]
			userEnd := strings.IndexAny(userPart, " <>\"\n\r")
			if userEnd == -1 {
				userEnd = len(userPart)
			}
			username := userPart[:userEnd]
			if len(username) > 3 && len(username) < 25 {
				metadata["user"] = username
			}
		}
	}
	
	// Extract comment link
	if strings.Contains(submissionInfo, "[comments]") {
		// Find the comment link
		commentStart := strings.Index(submissionInfo, `<a href="`)
		if commentStart != -1 {
			urlStart := commentStart + 9
			urlEnd := strings.Index(submissionInfo[urlStart:], `"`)
			if urlEnd != -1 {
				commentURL := submissionInfo[urlStart : urlStart+urlEnd]
				if strings.Contains(commentURL, "reddit.com") {
					metadata["comments_url"] = commentURL
				}
			}
		}
	}
	
	return metadata
}

// extractTextFromTableOnly extracts only text content from Reddit table without images
func extractTextFromTableOnly(content string) string {
	// Just extract text content and remove table structure
	text := extractTextFromTable(content)
	if text == content {
		// Fallback: remove table tags entirely
		formatted := strings.ReplaceAll(content, "<table>", "")
		formatted = strings.ReplaceAll(formatted, "</table>", "")
		formatted = strings.ReplaceAll(formatted, "<tr>", "")
		formatted = strings.ReplaceAll(formatted, "</tr>", "")
		formatted = strings.ReplaceAll(formatted, "<td>", "")
		formatted = strings.ReplaceAll(formatted, "</td>", "")
		return formatted
	}
	return text
}

// removeImagesFromContent removes img tags from content to prevent duplicates
func removeImagesFromContent(content string) string {
	// Remove img tags and any surrounding divs that contain only images
	formatted := content
	
	// Remove standalone img tags
	for strings.Contains(formatted, "<img") {
		start := strings.Index(formatted, "<img")
		if start == -1 {
			break
		}
		end := strings.Index(formatted[start:], ">")
		if end == -1 {
			break
		}
		formatted = formatted[:start] + formatted[start+end+1:]
	}
	
	// Remove links to image files (preview.redd.it, i.redd.it, etc.)
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	imageDomains := []string{"preview.redd.it", "i.redd.it", "imgur.com", "thumbs.redditmedia.com"}
	
	searchPos := 0
	for {
		start := strings.Index(formatted[searchPos:], "<a")
		if start == -1 {
			break
		}
		start += searchPos
		
		end := strings.Index(formatted[start:], "</a>")
		if end == -1 {
			break
		}
		linkContent := formatted[start : start+end+4]
		
		// Check if this link points to an image
		isImageLink := false
		for _, domain := range imageDomains {
			if strings.Contains(linkContent, domain) {
				isImageLink = true
				break
			}
		}
		if !isImageLink {
			for _, ext := range imageExtensions {
				if strings.Contains(linkContent, ext) {
					isImageLink = true
					break
				}
			}
		}
		
		if isImageLink {
			// Remove the entire link
			formatted = formatted[:start] + formatted[start+end+4:]
			// Continue searching from the same position since content shifted
		} else {
			// Move search position past this link
			searchPos = start + end + 4
		}
	}
	
	// Remove divs that only contain images (like reddit-post-with-image)
	formatted = strings.ReplaceAll(formatted, `<div class="reddit-post-with-image">`, "")
	formatted = strings.ReplaceAll(formatted, `<div class="reddit-post-text">`, "")
	
	// Clean up any empty or malformed tags
	formatted = strings.ReplaceAll(formatted, "</div></div>", "")
	
	return formatted
}

// extractRedditFlair extracts post flair like [D], [R], [P] from Reddit titles
func extractRedditFlair(title string) string {
	if strings.HasPrefix(title, "[") {
		end := strings.Index(title, "]")
		if end != -1 && end < 10 { // Reasonable flair length
			return title[:end+1]
		}
	}
	return ""
}

// extractRedditUser extracts Reddit username from author name or content
func extractRedditUser(authorName string) string {
	if strings.HasPrefix(authorName, "/u/") {
		return authorName
	}
	return ""
}

// extractRedditUserFromContent extracts Reddit user from content when not in author field
func extractRedditUserFromContent(content string) string {
	// Look for "/u/username" pattern in content
	userStart := strings.Index(content, "/u/")
	if userStart != -1 {
		// Find the end of the username (space, >, or end of string)
		usernamePart := content[userStart:]
		endChars := []string{" ", ">", "<", "\"", "'", "\n", "\r"}
		
		endPos := len(usernamePart)
		for _, endChar := range endChars {
			if pos := strings.Index(usernamePart, endChar); pos != -1 && pos < endPos {
				endPos = pos
			}
		}
		
		username := usernamePart[:endPos]
		if len(username) > 3 && len(username) < 25 { // Reasonable username length
			return username
		}
	}
	return ""
}

// extractRedditCommentCount extracts comment count from Reddit content
func extractRedditCommentCount(content string) string {
	// Look for "[comments]" links or "comments" text
	if strings.Contains(content, "[comments]") {
		return "Comments"
	}
	if strings.Contains(content, "comments") && strings.Contains(content, "reddit.com") {
		return "Discussion"
	}
	return ""
}

// NewsHandler displays the news page
func NewsHandler(w http.ResponseWriter, r *http.Request) {
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
			return
		}
		
		totalCount, err = db.GetNewsArticleCountWithSearchAndUserPrefs(db.DB, userID, category, language, search)
	} else {
		articles, err = db.GetNewsArticlesWithSearchAndUserPrefs(db.DB, userID, articlesPerPage, offset, category, language, "")
		if err != nil {
			http.Error(w, "Failed to load news articles", http.StatusInternalServerError)
			return
		}
		
		totalCount, err = db.GetNewsArticleCountWithSearchAndUserPrefs(db.DB, userID, category, language, "")
	}
	if err != nil {
		http.Error(w, "Failed to count news articles", http.StatusInternalServerError)
		return
	}

	// Get source count for the specific category (considering user preferences)
	sourceCount, err := db.GetActiveSourceCountByCategoryForUser(db.DB, userID, category)
	if err != nil {
		sourceCount = 0 // Continue with 0 count
	}

	// Get sources with user preferences for the template
	sources, err := db.GetNewsSourcesWithUserPrefs(db.DB, userID)
	if err != nil {
		sources = []db.NewsSourceWithUserPref{} // Continue with empty sources
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
		SourceCount: sourceCount,
		HasNext:     page < totalPages,
		HasPrev:     page > 1,
		NextPage:    page + 1,
		PrevPage:    page - 1,
		Sources:     sources,
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		// Return just the article content for HTMX requests
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		
		// Use enhanced template parsing for HTMX requests
		tmpl, err := template.New("news_articles.html").Funcs(template.FuncMap{
			"substr": func(s interface{}, start, length int) string {
				var str string
				switch v := s.(type) {
				case string:
					str = v
				case template.HTML:
					str = string(v)
				default:
					return ""
				}
				
				if start < 0 {
					start = 0
				}
				// Convert to runes to handle UTF-8 properly
				runes := []rune(str)
				if start >= len(runes) {
					return ""
				}
				end := start + length
				if end > len(runes) {
					end = len(runes)
				}
				return string(runes[start:end])
			},
			"substrHTML": func(s interface{}, start, length int) template.HTML {
				var str string
				switch v := s.(type) {
				case string:
					str = v
				case template.HTML:
					str = string(v)
				default:
					return template.HTML("")
				}
				
				if start < 0 {
					start = 0
				}
				// Convert to runes to handle UTF-8 properly
				runes := []rune(str)
				if start >= len(runes) {
					return template.HTML("")
				}
				end := start + length
				if end > len(runes) {
					end = len(runes)
				}
				return template.HTML(string(runes[start:end]))
			},
		}).ParseFiles("templates/partials/news_articles.html")
		if err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, "Template execution error", http.StatusInternalServerError)
		}
		return
	}

	// Render full page
	tmpl, err := template.New("base.html").Funcs(template.FuncMap{
		"substr": func(s interface{}, start, length int) string {
			var str string
			switch v := s.(type) {
			case string:
				str = v
			case template.HTML:
				str = string(v)
			default:
				return ""
			}
			
			if start < 0 {
				start = 0
			}
			// Convert to runes to handle UTF-8 properly
			runes := []rune(str)
			if start >= len(runes) {
				return ""
			}
			end := start + length
			if end > len(runes) {
				end = len(runes)
			}
			return string(runes[start:end])
		},
		"substrHTML": func(s interface{}, start, length int) template.HTML {
			var str string
			switch v := s.(type) {
			case string:
				str = v
			case template.HTML:
				str = string(v)
			default:
				return template.HTML("")
			}
			
			if start < 0 {
				start = 0
			}
			// Convert to runes to handle UTF-8 properly
			runes := []rune(str)
			if start >= len(runes) {
				return template.HTML("")
			}
			end := start + length
			if end > len(runes) {
				end = len(runes)
			}
			return template.HTML(string(runes[start:end]))
		},
	}).ParseFiles("templates/base.html", "templates/news.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
	}
}

// renderNewsArticles renders just the articles list for HTMX updates
func renderNewsArticles(w http.ResponseWriter, data NewsPageData) {
	
	// Add error recovery
	defer func() {
		if r := recover(); r != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}()
	
	// Render manually to avoid template issues
	fmt.Fprintf(w, `<div class="news-stats">
		<span class="article-count">%d articles found</span>
	</div>
	
	<div class="articles-grid">`, data.TotalCount)
	
	for _, article := range data.Articles {
		content := string(article.Content)
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
		return
	}

	err = tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
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
		return
	}

	// Check if this is a request from the news dashboard (via query param)
	isDashboard := r.URL.Query().Get("dashboard") == "true"
	
	// Get current page context for news refresh
	category := r.URL.Query().Get("category")
	language := r.URL.Query().Get("language")
	search := r.URL.Query().Get("search")
	
	if isDashboard {
		// Return dashboard-style toggle
		enabledClass := "enabled"
		statusText := "✓"
		if !newState {
			enabledClass = "disabled"
			statusText = "✗"
		}
		
		// Get source name for display
		sources, err := db.GetAllNewsSources(db.DB)
		if err != nil {
			http.Error(w, "Failed to load source info", http.StatusInternalServerError)
			return
		}
		
		var sourceName string
		for _, source := range sources {
			if source.ID == sourceID {
				sourceName = source.Name
				break
			}
		}
		
		// Set HX-Trigger header to trigger news refresh
		w.Header().Set("HX-Trigger", "refreshNews")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		
		fmt.Fprintf(w, `<span class="source-toggle %s" 
			hx-post="/news/sources/%s/user-toggle?dashboard=true&category=%s&language=%s&search=%s" 
			hx-trigger="click"
			hx-swap="outerHTML"
			data-source-id="%s">%s %s</span>`,
			enabledClass, sourceID, category, language, search, sourceID, sourceName, statusText)
	} else {
		// Return news sources page style button
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
	var extractedImageURL string
	
	// Process content based on category
	if article.Category == "hackernews" {
		// Format Hacker News content specially
		if fullContent != "" {
			fullContent = formatHackerNewsContent(fullContent)
		}
		if content != "" {
			content = formatHackerNewsContent(content)
		}
	} else if article.Category == "reddit" {
		// Extract image from Reddit content if no main image exists
		originalContent := stringPtrToString(article.Content)
		hasMainImage := stringPtrToString(article.ImageURL) != ""
		
		if !hasMainImage {
			extractedImageURL = extractFirstImageFromRedditContent(originalContent)
			if extractedImageURL != "" {
				hasMainImage = true
			}
		}
		
		// Format Reddit content and remove duplicate images
		hasImage := hasMainImage || stringPtrToString(article.ThumbnailURL) != ""
		if fullContent != "" {
			fullContent = formatRedditContentWithImageCheck(fullContent, hasImage)
		}
		if content != "" {
			content = formatRedditContentWithImageCheck(content, hasImage)
		}
	} else {
		// Process line breaks for display for other content
		if fullContent != "" {
			fullContent = strings.ReplaceAll(fullContent, "\r\n", "\n")
			fullContent = strings.ReplaceAll(fullContent, "\r", "\n")
		}
		if content != "" {
			content = strings.ReplaceAll(content, "\r\n", "\n")
			content = strings.ReplaceAll(content, "\r", "\n")
		}
	}
	
	// Determine final image URL (use extracted image if available, otherwise original)
	finalImageURL := stringPtrToString(article.ImageURL)
	if extractedImageURL != "" {
		finalImageURL = extractedImageURL
	}
	
	// Decode HTML entities in the final image URL
	if finalImageURL != "" {
		finalImageURL = html.UnescapeString(finalImageURL)
	}
	
	templateArticle := TemplateArticle{
		ID:           article.ID.String(),
		Title:        article.Title,
		Content:      template.HTML(content),
		Summary:      stringPtrToString(article.Summary),
		FullContent:  template.HTML(fullContent),
		OriginalURL:  article.OriginalURL,
		ImageURL:     finalImageURL,
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
		return
	}

	err = tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
	}
}

// GetSectionSourcesHandler returns sources for a specific section with user preferences
func GetSectionSourcesHandler(w http.ResponseWriter, r *http.Request) {
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

	// Get section from query parameter
	section := r.URL.Query().Get("section")
	if section == "" {
		http.Error(w, "Section parameter required", http.StatusBadRequest)
		return
	}
	
	// Get current category, language, and search for news refresh
	category := r.URL.Query().Get("category")
	language := r.URL.Query().Get("language")
	search := r.URL.Query().Get("search")

	// Get sources for the section
	sources, err := db.GetAllNewsSources(db.DB)
	if err != nil {
		http.Error(w, "Failed to load sources", http.StatusInternalServerError)
		return
	}

	// Get user preferences
	userPrefs, err := db.GetUserNewsPreferences(db.DB, userID)
	if err != nil {
		http.Error(w, "Failed to load user preferences", http.StatusInternalServerError)
		return
	}

	// Filter sources by section and build response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	for _, source := range sources {
		if source.Category == section {
			isEnabled := true // Default to enabled
			if pref, exists := userPrefs[source.ID]; exists {
				isEnabled = pref
			}

			enabledClass := "enabled"
			statusText := "✓"
			if !isEnabled {
				enabledClass = "disabled"
				statusText = "✗"
			}

			fmt.Fprintf(w, `<span class="source-toggle %s" 
				hx-post="/news/sources/%s/user-toggle?dashboard=true&category=%s&language=%s&search=%s" 
				hx-trigger="click"
				hx-swap="outerHTML"
				hx-on::after-request="htmx.ajax('GET', '/news?page=1&category=%s&language=%s&search=%s', {target: '#news-content', swap: 'innerHTML'})"
				data-source-id="%s">%s %s</span>`,
				enabledClass, source.ID.String(), category, language, search, category, language, search, source.ID.String(), source.Name, statusText)
		}
	}
}