package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"ohabits.com/internal/db"
)

const (
	JikanBaseURL       = "https://api.jikan.moe/v4"
	JikanRateLimit     = 500 * time.Millisecond // More conservative: 2 requests per second
	JikanMaxRetries    = 3                      // Maximum retry attempts for rate limit errors
	JikanRetryBaseWait = 2 * time.Second        // Base wait time for exponential backoff
)

// ProgressCallback is a function type for progress updates
type ProgressCallback func(message string, current, total int)

// jikanRequest makes a rate-limited HTTP request with retry logic for 429 errors
func jikanRequest(url string) (*http.Response, error) {
	for attempt := 0; attempt <= JikanMaxRetries; attempt++ {
		// Rate limiting
		time.Sleep(JikanRateLimit)
		
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		
		// If not rate limited, return the response
		if resp.StatusCode != 429 {
			return resp, nil
		}
		
		resp.Body.Close()
		
		// If this was our last attempt, return the error
		if attempt == JikanMaxRetries {
			return nil, fmt.Errorf("Jikan API rate limit exceeded after %d retries", JikanMaxRetries)
		}
		
		// Exponential backoff: wait longer after each failed attempt
		waitTime := JikanRetryBaseWait * time.Duration(1<<uint(attempt))
		time.Sleep(waitTime)
	}
	
	return nil, fmt.Errorf("unexpected end of retry loop")
}

// Jikan API Response structs
type JikanAnimeSearchResponse struct {
	Data       []db.JikanAnime   `json:"data"`
	Pagination JikanPagination `json:"pagination"`
}

type JikanDateRange struct {
	From   *string `json:"from"`
	To     *string `json:"to"`
	String string  `json:"string"`
}

type JikanBroadcast struct {
	Day      *string `json:"day"`
	Time     *string `json:"time"`
	Timezone *string `json:"timezone"`
	String   *string `json:"string"`
}

type JikanPagination struct {
	LastVisiblePage int  `json:"last_visible_page"`
	HasNextPage     bool `json:"has_next_page"`
	CurrentPage     int  `json:"current_page"`
	Items           struct {
		Count   int `json:"count"`
		Total   int `json:"total"`
		PerPage int `json:"per_page"`
	} `json:"items"`
}

// SearchAnime searches for anime using the Jikan API
func SearchAnime(query string) ([]db.JikanAnime, error) {
	if query == "" {
		return []db.JikanAnime{}, nil
	}

	encodedQuery := url.QueryEscape(query)
	apiURL := fmt.Sprintf("%s/anime?q=%s&limit=10", JikanBaseURL, encodedQuery)

	resp, err := jikanRequest(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to search anime: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Jikan API returned status %d", resp.StatusCode)
	}

	var searchResp JikanAnimeSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode Jikan response: %w", err)
	}

	return searchResp.Data, nil
}

// GetAnimeByID gets anime details by MAL ID
func GetAnimeByID(malID int) (*db.JikanAnime, error) {
	apiURL := fmt.Sprintf("%s/anime/%d", JikanBaseURL, malID)

	resp, err := jikanRequest(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get anime: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Jikan API returned status %d", resp.StatusCode)
	}

	var animeResp struct {
		Data db.JikanAnime `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&animeResp); err != nil {
		return nil, fmt.Errorf("failed to decode anime response: %w", err)
	}

	return &animeResp.Data, nil
}

// GetAnimeEpisodes gets episodes for an anime using the Jikan API with pagination
func GetAnimeEpisodes(malID int) ([]db.JikanEpisode, error) {
	return GetAnimeEpisodesWithProgress(malID, nil)
}

// GetAnimeEpisodesWithProgress gets episodes with progress callback
func GetAnimeEpisodesWithProgress(malID int, progressCallback ProgressCallback) ([]db.JikanEpisode, error) {
	var allEpisodes []db.JikanEpisode
	page := 1
	totalPages := 0
	
	for {
		apiURL := fmt.Sprintf("%s/anime/%d/episodes?page=%d", JikanBaseURL, malID, page)

		// Send progress update
		if progressCallback != nil {
			if totalPages > 0 {
				progressCallback(fmt.Sprintf("Fetching episodes (page %d of %d)...", page, totalPages), page, totalPages)
			} else {
				progressCallback(fmt.Sprintf("Fetching episodes (page %d)...", page), page, -1)
			}
		}

		resp, err := jikanRequest(apiURL)
		if err != nil {
			return nil, fmt.Errorf("failed to get anime episodes page %d: %w", page, err)
		}

		if resp.StatusCode == 404 {
			if page == 1 {
				resp.Body.Close()
				return []db.JikanEpisode{}, nil // Anime exists but no episodes
			}
			// No more pages
			resp.Body.Close()
			break
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("Jikan API returned status %d for page %d", resp.StatusCode, page)
		}

		var episodesResp struct {
			Data       []db.JikanEpisode `json:"data"`
			Pagination JikanPagination   `json:"pagination"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&episodesResp); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode episodes response page %d: %w", page, err)
		}
		resp.Body.Close()

		// Update total pages if we just learned it
		if totalPages == 0 {
			totalPages = episodesResp.Pagination.LastVisiblePage
		}

		// Add episodes from this page
		allEpisodes = append(allEpisodes, episodesResp.Data...)

		// Send completion update for this page
		if progressCallback != nil {
			progressCallback(fmt.Sprintf("Fetched %d episodes so far...", len(allEpisodes)), page, totalPages)
		}

		// Check if there are more pages
		if !episodesResp.Pagination.HasNextPage {
			break
		}
		
		page++
		
		// Safety limit to prevent infinite loops (max 50 pages = ~1250 episodes)
		if page > 50 {
			break
		}
	}

	// Send final completion message
	if progressCallback != nil {
		progressCallback(fmt.Sprintf("Successfully fetched %d episodes!", len(allEpisodes)), page, totalPages)
	}

	return allEpisodes, nil
}

// GetAnimeEpisodeVideos gets episode images for an anime using the Jikan API
func GetAnimeEpisodeVideos(malID int) (map[int]string, error) {
	return GetAnimeEpisodeVideosWithProgress(malID, nil)
}

// GetAnimeEpisodeVideosWithProgress gets episode images with progress callback
func GetAnimeEpisodeVideosWithProgress(malID int, progressCallback ProgressCallback) (map[int]string, error) {
	episodeImages := make(map[int]string)
	page := 1
	totalPages := 0
	
	for {
		apiURL := fmt.Sprintf("%s/anime/%d/videos/episodes?page=%d", JikanBaseURL, malID, page)

		// Send progress update
		if progressCallback != nil {
			if totalPages > 0 {
				progressCallback(fmt.Sprintf("Fetching episode images (page %d of %d)...", page, totalPages), page, totalPages)
			} else {
				progressCallback(fmt.Sprintf("Fetching episode images (page %d)...", page), page, -1)
			}
		}

		resp, err := jikanRequest(apiURL)
		if err != nil {
			return nil, fmt.Errorf("failed to get anime episode videos page %d: %w", page, err)
		}

		if resp.StatusCode == 404 {
			resp.Body.Close()
			break // No more pages or no videos
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("Jikan API returned status %d for episode videos page %d", resp.StatusCode, page)
		}

		var videosResp struct {
			Data       []db.JikanEpisodeVideo `json:"data"`
			Pagination JikanPagination        `json:"pagination"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&videosResp); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode episode videos response page %d: %w", page, err)
		}
		resp.Body.Close()

		// Update total pages if we just learned it
		if totalPages == 0 {
			totalPages = videosResp.Pagination.LastVisiblePage
		}

		// Map episode MAL IDs to image URLs
		for _, video := range videosResp.Data {
			if video.Images.JPG.ImageURL != "" {
				episodeImages[video.MalID] = video.Images.JPG.ImageURL
			}
		}

		// Send progress update
		if progressCallback != nil {
			progressCallback(fmt.Sprintf("Found %d episode images so far...", len(episodeImages)), page, totalPages)
		}

		// Check if there are more pages
		if !videosResp.Pagination.HasNextPage {
			break
		}
		
		page++
		
		// Safety limit to prevent infinite loops
		if page > 50 {
			break
		}
	}

	// Send final completion message
	if progressCallback != nil {
		progressCallback(fmt.Sprintf("Successfully found %d episode images!", len(episodeImages)), page, totalPages)
	}

	return episodeImages, nil
}

// ConvertJikanAnimeToShow converts a Jikan anime to our internal Show model
func ConvertJikanAnimeToShow(anime db.JikanAnime) db.Show {
	// Convert Jikan status to TV show compatible status
	var status string
	switch anime.Status {
	case "Finished Airing":
		status = "Ended"
	case "Currently Airing":
		status = "Running"
	case "Not yet aired":
		status = "TBD"
	case "Cancelled":
		status = "Ended" // Treat cancelled anime as ended
	case "Hiatus":
		status = "Running" // Treat hiatus as still running
	default:
		status = anime.Status // Fallback to original status
	}
	
	show := db.Show{
		ExternalID: anime.MalID, // MAL ID for anime
		ShowType:   "anime",
		Name:       anime.Title,
		Summary:    anime.Synopsis,
		Status:     &status,
		TVMazeID:   anime.MalID, // Legacy field for backward compatibility
	}

	// Handle image URL - prefer JPG large image
	if anime.Images.JPG.LargeImageURL != "" {
		show.ImageURL = &anime.Images.JPG.LargeImageURL
	} else if anime.Images.JPG.ImageURL != "" {
		show.ImageURL = &anime.Images.JPG.ImageURL
	}

	// Handle year (convert to premiered date)
	if anime.Year != nil {
		premiered := time.Date(*anime.Year, 1, 1, 0, 0, 0, 0, time.UTC)
		show.Premiered = &premiered
	}

	// Handle studios (treat as network)
	if len(anime.Studios) > 0 {
		show.Network = &anime.Studios[0].Name
	}

	// Handle genres
	if len(anime.Genres) > 0 {
		var genreNames []string
		for _, genre := range anime.Genres {
			genreNames = append(genreNames, genre.Name)
		}
		genresJSON, _ := json.Marshal(genreNames)
		show.Genres = genresJSON
	}

	// Handle rating/score
	if anime.Score != nil {
		ratingData := map[string]interface{}{
			"average": *anime.Score,
		}
		ratingJSON, _ := json.Marshal(ratingData)
		show.Rating = ratingJSON
	}

	return show
}

// ConvertJikanEpisodeToEpisode converts a Jikan episode to our internal Episode model
func ConvertJikanEpisodeToEpisode(jikanEpisode db.JikanEpisode, showID uuid.UUID, userID uuid.UUID, imageURL *string) db.Episode {
	episode := db.Episode{
		ShowID:     showID,
		UserID:     userID,
		ExternalID: jikanEpisode.MalID, // MAL episode ID
		ShowType:   "anime",
		Name:       jikanEpisode.Title,
		Season:     1,                   // Anime episodes typically don't have seasons in the same way as TV shows
		Number:     jikanEpisode.MalID,  // MAL ID serves as the episode number for anime
		Summary:    nil,                 // Jikan episodes don't have synopsis in the episodes endpoint
		Runtime:    nil,                 // Duration not provided in episodes endpoint
		ImageURL:   imageURL,            // Episode image from videos API
		Filler:     jikanEpisode.Filler, // Filler episode indicator
		Recap:      jikanEpisode.Recap,  // Recap episode indicator
	}

	// Handle air date - Jikan provides dates in ISO format
	if jikanEpisode.Aired != nil && *jikanEpisode.Aired != "" {
		// Try parsing the ISO date format from Jikan
		if airDate, err := time.Parse(time.RFC3339, *jikanEpisode.Aired); err == nil {
			episode.AirDate = &airDate
		} else if airDate, err := time.Parse("2006-01-02T15:04:05", *jikanEpisode.Aired); err == nil {
			episode.AirDate = &airDate
		} else if airDate, err := time.Parse("2006-01-02", *jikanEpisode.Aired); err == nil {
			episode.AirDate = &airDate
		}
	}

	return episode
}