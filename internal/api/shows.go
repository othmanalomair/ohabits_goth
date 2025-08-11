package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"ohabits.com/internal/db"
)

const (
	TVMazeBaseURL = "https://api.tvmaze.com"
)

// SearchShows searches for shows using TVmaze API
func SearchShows(query string) ([]db.TVMazeSearchResult, error) {
	if strings.TrimSpace(query) == "" {
		return []db.TVMazeSearchResult{}, nil
	}

	encodedQuery := url.QueryEscape(query)
	apiURL := fmt.Sprintf("%s/search/shows?q=%s", TVMazeBaseURL, encodedQuery)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to search shows: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var results []db.TVMazeSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	return results, nil
}

// GetShowByID gets show details by TVmaze ID
func GetShowByID(tvmazeID int) (*db.TVMazeShow, error) {
	apiURL := fmt.Sprintf("%s/shows/%d", TVMazeBaseURL, tvmazeID)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get show: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var show db.TVMazeShow
	if err := json.NewDecoder(resp.Body).Decode(&show); err != nil {
		return nil, fmt.Errorf("failed to decode show response: %w", err)
	}

	return &show, nil
}

// GetShowEpisodes gets episodes for a show using TVmaze API
func GetShowEpisodes(tvmazeID int) ([]db.TVMazeEpisode, error) {
	apiURL := fmt.Sprintf("%s/shows/%d/episodes", TVMazeBaseURL, tvmazeID)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get episodes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var episodes []db.TVMazeEpisode
	if err := json.NewDecoder(resp.Body).Decode(&episodes); err != nil {
		return nil, fmt.Errorf("failed to decode episodes response: %w", err)
	}

	return episodes, nil
}

// ConvertTVMazeShowToShow converts a TVMaze show to our internal Show model
func ConvertTVMazeShowToShow(tvShow db.TVMazeShow) db.Show {
	show := db.Show{
		TVMazeID: tvShow.ID,
		Name:     tvShow.Name,
		Summary:  tvShow.Summary,
		Status:   &tvShow.Status,
	}

	// Handle image URL
	if tvShow.Image != nil && tvShow.Image.Medium != nil {
		show.ImageURL = tvShow.Image.Medium
	}

	// Handle premiered date
	if tvShow.Premiered != nil {
		if premiered, err := time.Parse("2006-01-02", *tvShow.Premiered); err == nil {
			show.Premiered = &premiered
		}
	}

	// Handle ended date
	if tvShow.Ended != nil {
		if ended, err := time.Parse("2006-01-02", *tvShow.Ended); err == nil {
			show.Ended = &ended
		}
	}

	// Handle network
	if tvShow.Network != nil {
		show.Network = &tvShow.Network.Name
	}

	// Handle genres
	if len(tvShow.Genres) > 0 {
		genresJSON, _ := json.Marshal(tvShow.Genres)
		show.Genres = genresJSON
	}

	// Handle rating
	if tvShow.Rating != nil {
		ratingJSON, _ := json.Marshal(tvShow.Rating)
		show.Rating = ratingJSON
	}

	return show
}

// ConvertTVMazeEpisodeToEpisode converts a TVMaze episode to our internal Episode model
func ConvertTVMazeEpisodeToEpisode(tvEpisode db.TVMazeEpisode, showID string, userID string) db.Episode {
	episode := db.Episode{
		TVMazeID: tvEpisode.ID,
		Name:     tvEpisode.Name,
		Season:   tvEpisode.Season,
		Number:   tvEpisode.Number,
		Summary:  tvEpisode.Summary,
		Runtime:  tvEpisode.Runtime,
	}

	// Parse showID and userID
	if showUUID, err := parseUUID(showID); err == nil {
		episode.ShowID = showUUID
	}
	if userUUID, err := parseUUID(userID); err == nil {
		episode.UserID = userUUID
	}

	// Handle air date
	if tvEpisode.AirDate != nil {
		if airDate, err := time.Parse("2006-01-02", *tvEpisode.AirDate); err == nil {
			episode.AirDate = &airDate
		}
	}

	// Handle image URL
	if tvEpisode.Image != nil && tvEpisode.Image.Medium != nil {
		episode.ImageURL = tvEpisode.Image.Medium
	}

	return episode
}

// Helper function to parse UUID strings
func parseUUID(uuidStr string) (uuid.UUID, error) {
	return uuid.Parse(uuidStr)
}

// StripHTMLTags removes HTML tags from text (simple implementation for summaries)
func StripHTMLTags(text string) string {
	if text == "" {
		return ""
	}
	
	// Simple regex to remove HTML tags - for production, use a proper HTML parser
	text = strings.ReplaceAll(text, "<p>", "")
	text = strings.ReplaceAll(text, "</p>", "\n")
	text = strings.ReplaceAll(text, "<br>", "\n")
	text = strings.ReplaceAll(text, "<br/>", "\n")
	text = strings.ReplaceAll(text, "<b>", "")
	text = strings.ReplaceAll(text, "</b>", "")
	text = strings.ReplaceAll(text, "<i>", "")
	text = strings.ReplaceAll(text, "</i>", "")
	
	return strings.TrimSpace(text)
}