package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"ohabits.com/internal/api"
	"ohabits.com/internal/db"
)

type SyncService struct {
	db          *pgxpool.Pool
	isRunning   bool
	stopChannel chan bool
	mutex       sync.RWMutex
}

type SyncSettings struct {
	EpisodeSyncIntervalHours int
	InfoSyncIntervalHours    int
	MaxConcurrentSyncs       int
}

type SyncResult struct {
	ShowID         uuid.UUID
	SyncType       string
	Status         string
	EpisodesAdded  int
	InfoUpdated    bool
	ErrorMessage   string
	SyncDurationMs int64
}

// NewSyncService creates a new sync service instance
func NewSyncService(db *pgxpool.Pool) *SyncService {
	return &SyncService{
		db:          db,
		isRunning:   false,
		stopChannel: make(chan bool),
	}
}

// Start begins the sync service with periodic updates
func (s *SyncService) Start() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isRunning {
		return
	}

	s.isRunning = true
	go s.syncLoop()
	log.Println("Sync service started")
}

// Stop halts the sync service
func (s *SyncService) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.isRunning {
		return
	}

	s.isRunning = false
	s.stopChannel <- true
	log.Println("Sync service stopped")
}

// syncLoop runs the main sync loop
func (s *SyncService) syncLoop() {
	// Initial delay to avoid startup spam
	time.Sleep(30 * time.Second)

	// Check every hour
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChannel:
			return
		case <-ticker.C:
			s.performSync()
		}
	}
}

// performSync checks for shows that need updating and syncs them
func (s *SyncService) performSync() {
	log.Println("Starting periodic sync check...")

	settings, err := s.getSyncSettings()
	if err != nil {
		log.Printf("Error getting sync settings: %v", err)
		return
	}

	// Get shows that need episode sync
	showsForEpisodeSync, err := s.getShowsNeedingEpisodeSync(settings.EpisodeSyncIntervalHours)
	if err != nil {
		log.Printf("Error getting shows for episode sync: %v", err)
		return
	}

	// Get shows that need info sync
	showsForInfoSync, err := s.getShowsNeedingInfoSync(settings.InfoSyncIntervalHours)
	if err != nil {
		log.Printf("Error getting shows for info sync: %v", err)
		return
	}

	if len(showsForEpisodeSync) == 0 && len(showsForInfoSync) == 0 {
		log.Println("No shows need syncing at this time")
		return
	}

	log.Printf("Found %d shows for episode sync, %d shows for info sync", 
		len(showsForEpisodeSync), len(showsForInfoSync))

	// Create semaphore for concurrent sync limit
	semaphore := make(chan struct{}, settings.MaxConcurrentSyncs)
	var wg sync.WaitGroup

	// Sync episodes
	for _, show := range showsForEpisodeSync {
		wg.Add(1)
		go func(show db.Show) {
			defer wg.Done()
			semaphore <- struct{}{} // Acquire
			defer func() { <-semaphore }() // Release

			result := s.syncEpisodes(show)
			s.logSyncResult(result)
		}(show)
	}

	// Sync show info
	for _, show := range showsForInfoSync {
		wg.Add(1)
		go func(show db.Show) {
			defer wg.Done()
			semaphore <- struct{}{} // Acquire
			defer func() { <-semaphore }() // Release

			result := s.syncShowInfo(show)
			s.logSyncResult(result)
		}(show)
	}

	wg.Wait()
	
	// Cleanup old notifications (run once per sync cycle)
	if err := db.CleanupOldNotifications(s.db); err != nil {
		log.Printf("Failed to cleanup old notifications: %v", err)
	}
	
	log.Println("Sync check completed")
}

// getSyncSettings retrieves sync configuration from database
func (s *SyncService) getSyncSettings() (*SyncSettings, error) {
	settings := &SyncSettings{
		EpisodeSyncIntervalHours: 6,  // default
		InfoSyncIntervalHours:    24, // default
		MaxConcurrentSyncs:       3,  // default
	}

	rows, err := s.db.Query(context.Background(), 
		"SELECT setting_key, setting_value FROM sync_settings WHERE setting_key IN ($1, $2, $3)",
		"episode_sync_interval_hours", "info_sync_interval_hours", "max_concurrent_syncs")
	if err != nil {
		return settings, nil // Use defaults on error
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}

		switch key {
		case "episode_sync_interval_hours":
			if val, err := time.ParseDuration(value + "h"); err == nil {
				settings.EpisodeSyncIntervalHours = int(val.Hours())
			}
		case "info_sync_interval_hours":
			if val, err := time.ParseDuration(value + "h"); err == nil {
				settings.InfoSyncIntervalHours = int(val.Hours())
			}
		case "max_concurrent_syncs":
			if val, err := time.ParseDuration(value + "s"); err == nil {
				settings.MaxConcurrentSyncs = int(val.Seconds())
			}
		}
	}

	return settings, nil
}

// getShowsNeedingEpisodeSync gets shows that need episode updates
func (s *SyncService) getShowsNeedingEpisodeSync(intervalHours int) ([]db.Show, error) {
	query := `
		SELECT id, user_id, tvmaze_id, name, summary, image_url, status, premiered, ended, network, genres, rating, created_at, updated_at
		FROM shows 
		WHERE (last_episode_sync IS NULL OR last_episode_sync < NOW() - INTERVAL '%d hours')
		AND status != 'Ended'
		ORDER BY last_episode_sync ASC NULLS FIRST
		LIMIT 20`

	rows, err := s.db.Query(context.Background(), fmt.Sprintf(query, intervalHours))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shows []db.Show
	for rows.Next() {
		var show db.Show
		err := rows.Scan(&show.ID, &show.UserID, &show.TVMazeID, &show.Name, &show.Summary,
			&show.ImageURL, &show.Status, &show.Premiered, &show.Ended, &show.Network,
			&show.Genres, &show.Rating, &show.CreatedAt, &show.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning show: %v", err)
			continue
		}
		shows = append(shows, show)
	}

	return shows, nil
}

// getShowsNeedingInfoSync gets shows that need info updates
func (s *SyncService) getShowsNeedingInfoSync(intervalHours int) ([]db.Show, error) {
	query := `
		SELECT id, user_id, tvmaze_id, name, summary, image_url, status, premiered, ended, network, genres, rating, created_at, updated_at
		FROM shows 
		WHERE (last_info_sync IS NULL OR last_info_sync < NOW() - INTERVAL '%d hours')
		ORDER BY last_info_sync ASC NULLS FIRST
		LIMIT 10`

	rows, err := s.db.Query(context.Background(), fmt.Sprintf(query, intervalHours))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shows []db.Show
	for rows.Next() {
		var show db.Show
		err := rows.Scan(&show.ID, &show.UserID, &show.TVMazeID, &show.Name, &show.Summary,
			&show.ImageURL, &show.Status, &show.Premiered, &show.Ended, &show.Network,
			&show.Genres, &show.Rating, &show.CreatedAt, &show.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning show: %v", err)
			continue
		}
		shows = append(shows, show)
	}

	return shows, nil
}

// syncEpisodes synchronizes episodes for a show
func (s *SyncService) syncEpisodes(show db.Show) SyncResult {
	startTime := time.Now()
	result := SyncResult{
		ShowID:   show.ID,
		SyncType: "episodes",
		Status:   "success",
	}

	// Fetch episodes from API
	apiEpisodes, err := api.GetAllEpisodes(show.TVMazeID)
	if err != nil {
		result.Status = "error"
		result.ErrorMessage = err.Error()
		result.SyncDurationMs = time.Since(startTime).Milliseconds()
		return result
	}

	// Get existing episodes
	existingEpisodes, err := db.GetEpisodesByShow(s.db, show.ID, show.UserID)
	if err != nil {
		result.Status = "error"
		result.ErrorMessage = err.Error()
		result.SyncDurationMs = time.Since(startTime).Milliseconds()
		return result
	}

	// Create map of existing episodes by TVMaze ID
	existingEpisodesMap := make(map[int]bool)
	for _, ep := range existingEpisodes {
		existingEpisodesMap[ep.TVMazeID] = true
	}

	// Find new episodes to add
	var newEpisodes []db.Episode
	for _, apiEpisode := range apiEpisodes {
		if !existingEpisodesMap[apiEpisode.ID] {
			episode := api.ConvertTVMazeEpisodeToEpisode(apiEpisode, show.ID, show.UserID)
			newEpisodes = append(newEpisodes, episode)
		}
	}

	// Add new episodes to database
	if len(newEpisodes) > 0 {
		if err := db.CreateEpisodes(s.db, newEpisodes); err != nil {
			result.Status = "error"
			result.ErrorMessage = err.Error()
		} else {
			result.EpisodesAdded = len(newEpisodes)
			log.Printf("Added %d new episodes for show: %s", len(newEpisodes), show.Name)
			
			// Create notification for new episodes
			if err := db.CreateEpisodeNotification(s.db, show.UserID, show.Name, len(newEpisodes), show.ID); err != nil {
				log.Printf("Failed to create notification for new episodes: %v", err)
			}
		}
	}
	
	// Add testing notifications (simulating new episodes from different shows)
	s.addTestNotifications(show.UserID)

	// Update last sync time
	s.updateLastEpisodeSync(show.ID)

	result.SyncDurationMs = time.Since(startTime).Milliseconds()
	return result
}

// syncShowInfo synchronizes show information
func (s *SyncService) syncShowInfo(show db.Show) SyncResult {
	startTime := time.Now()
	result := SyncResult{
		ShowID:   show.ID,
		SyncType: "info",
		Status:   "success",
	}

	// Fetch show info from API
	apiShow, err := api.GetShowDetails(show.TVMazeID)
	if err != nil {
		result.Status = "error"
		result.ErrorMessage = err.Error()
		result.SyncDurationMs = time.Since(startTime).Milliseconds()
		return result
	}

	// Check if status or other info changed
	statusChanged := false
	if show.Status == nil || *show.Status != apiShow.Status {
		statusChanged = true
	}

	if statusChanged {
		// Update show info in database
		updatedShow := api.ConvertTVMazeShowToShow(*apiShow)
		updatedShow.ID = show.ID
		updatedShow.UserID = show.UserID

		if err := s.updateShowInfo(updatedShow); err != nil {
			result.Status = "error"
			result.ErrorMessage = err.Error()
		} else {
			result.InfoUpdated = true
			log.Printf("Updated info for show: %s (status: %s)", show.Name, apiShow.Status)
		}
	}

	// Update last sync time
	s.updateLastInfoSync(show.ID)

	result.SyncDurationMs = time.Since(startTime).Milliseconds()
	return result
}

// updateLastEpisodeSync updates the last episode sync timestamp
func (s *SyncService) updateLastEpisodeSync(showID uuid.UUID) {
	_, err := s.db.Exec(context.Background(),
		"UPDATE shows SET last_episode_sync = NOW() WHERE id = $1", showID)
	if err != nil {
		log.Printf("Error updating last episode sync: %v", err)
	}
}

// updateLastInfoSync updates the last info sync timestamp
func (s *SyncService) updateLastInfoSync(showID uuid.UUID) {
	_, err := s.db.Exec(context.Background(),
		"UPDATE shows SET last_info_sync = NOW() WHERE id = $1", showID)
	if err != nil {
		log.Printf("Error updating last info sync: %v", err)
	}
}

// updateShowInfo updates show information in the database
func (s *SyncService) updateShowInfo(show db.Show) error {
	_, err := s.db.Exec(context.Background(), `
		UPDATE shows SET 
			name = $1, summary = $2, image_url = $3, status = $4, 
			premiered = $5, ended = $6, network = $7, genres = $8, rating = $9,
			updated_at = NOW()
		WHERE id = $10`,
		show.Name, show.Summary, show.ImageURL, show.Status, show.Premiered,
		show.Ended, show.Network, show.Genres, show.Rating, show.ID)
	return err
}

// logSyncResult logs the sync operation result
func (s *SyncService) logSyncResult(result SyncResult) {
	_, err := s.db.Exec(context.Background(), `
		INSERT INTO sync_logs (show_id, sync_type, status, episodes_added, info_updated, error_message, sync_duration_ms)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		result.ShowID, result.SyncType, result.Status, result.EpisodesAdded,
		result.InfoUpdated, result.ErrorMessage, result.SyncDurationMs)
	if err != nil {
		log.Printf("Error logging sync result: %v", err)
	}
}

// addTestNotifications adds testing notifications to simulate new episodes
func (s *SyncService) addTestNotifications(userID uuid.UUID) {
	testShows := []struct {
		name         string
		episodeCount int
	}{
		{"Attack on Titan", 2},
		{"One Piece", 3},
		{"Breaking Bad", 1},
		{"The Mandalorian", 1},
		{"Stranger Things", 4},
	}
	
	for _, show := range testShows {
		// Generate a random show ID for testing
		showID := uuid.New()
		
		// Only add notifications occasionally to avoid spam
		if time.Now().Unix()%7 == 0 { // Every 7th call approximately
			if err := db.CreateEpisodeNotification(s.db, userID, show.name, show.episodeCount, showID); err != nil {
				log.Printf("Failed to create test notification for %s: %v", show.name, err)
			}
		}
	}
}