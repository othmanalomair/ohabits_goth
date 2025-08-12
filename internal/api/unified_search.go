package api

import (
	"log"
	"sync"

	"ohabits.com/internal/db"
)

// UnifiedSearch performs search across both TVMaze and Jikan APIs
func UnifiedSearch(query string) ([]db.UnifiedSearchResult, error) {
	if query == "" {
		return []db.UnifiedSearchResult{}, nil
	}

	var wg sync.WaitGroup
	var tvResults []db.TVMazeSearchResult
	var animeResults []db.JikanAnime
	var tvErr, animeErr error

	// Search TV shows using TVMaze API
	wg.Add(1)
	go func() {
		defer wg.Done()
		tvResults, tvErr = SearchShows(query)
	}()

	// Search anime using Jikan API
	wg.Add(1)
	go func() {
		defer wg.Done()
		animeResults, animeErr = SearchAnime(query)
	}()

	// Wait for both searches to complete
	wg.Wait()

	// Combine results
	var unifiedResults []db.UnifiedSearchResult

	// Add TV show results
	if tvErr != nil {
		log.Printf("TVMaze search error: %v", tvErr)
	} else {
		for _, tvResult := range tvResults {
			unifiedResults = append(unifiedResults, db.UnifiedSearchResult{
				Type:   "tv",
				TVShow: &tvResult,
			})
		}
	}

	// Add anime results
	if animeErr != nil {
		log.Printf("Jikan search error: %v", animeErr)
	} else {
		for _, animeResult := range animeResults {
			unifiedResults = append(unifiedResults, db.UnifiedSearchResult{
				Type:  "anime",
				Anime: &animeResult,
			})
		}
	}

	return unifiedResults, nil
}

// Helper function to get external ID from unified result
func GetExternalIDFromResult(result db.UnifiedSearchResult) int {
	if result.Type == "tv" && result.TVShow != nil {
		return result.TVShow.Show.ID
	} else if result.Type == "anime" && result.Anime != nil {
		return result.Anime.MalID
	}
	return 0
}

// Helper function to get title from unified result
func GetTitleFromResult(result db.UnifiedSearchResult) string {
	if result.Type == "tv" && result.TVShow != nil {
		return result.TVShow.Show.Name
	} else if result.Type == "anime" && result.Anime != nil {
		return result.Anime.Title
	}
	return ""
}

// Helper function to convert unified result to Show model
func ConvertUnifiedResultToShow(result db.UnifiedSearchResult) db.Show {
	if result.Type == "tv" && result.TVShow != nil {
		return ConvertTVMazeShowToShow(result.TVShow.Show)
	} else if result.Type == "anime" && result.Anime != nil {
		return ConvertJikanAnimeToShow(*result.Anime)
	}
	return db.Show{}
}