package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Show database functions
func GetAllShows(db *pgxpool.Pool, userID uuid.UUID) ([]Show, error) {
	shows := []Show{}

	rows, err := db.Query(context.Background(), `
		SELECT id, user_id, tvmaze_id, name, summary, image_url, status, premiered, ended, network, genres, rating, created_at, updated_at 
		FROM shows 
		WHERE user_id = $1
		ORDER BY name
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var show Show
		err := rows.Scan(&show.ID, &show.UserID, &show.TVMazeID, &show.Name, &show.Summary, &show.ImageURL, &show.Status, &show.Premiered, &show.Ended, &show.Network, &show.Genres, &show.Rating, &show.CreatedAt, &show.UpdatedAt)
		if err != nil {
			return nil, err
		}
		shows = append(shows, show)
	}
	return shows, nil
}

// GetAllShowsWithEpisodeCounts gets all shows for a user with episode and watched counts
func GetAllShowsWithEpisodeCounts(db *pgxpool.Pool, userID uuid.UUID) ([]Show, error) {
	shows := []Show{}

	rows, err := db.Query(context.Background(), `
		SELECT 
			s.id, s.user_id, s.tvmaze_id, s.name, s.summary, s.image_url, s.status, 
			s.premiered, s.ended, s.network, s.genres, s.rating, s.created_at, s.updated_at,
			COALESCE(ep_counts.total_episodes, 0) as total_episodes,
			COALESCE(ep_counts.watched_episodes, 0) as watched_episodes
		FROM shows s
		LEFT JOIN (
			SELECT 
				e.show_id,
				COUNT(*) as total_episodes,
				COUNT(CASE WHEN et.watched = true THEN 1 END) as watched_episodes
			FROM episodes e
			LEFT JOIN episode_tracking et ON e.id = et.episode_id
			WHERE e.user_id = $1
			GROUP BY e.show_id
		) ep_counts ON s.id = ep_counts.show_id
		WHERE s.user_id = $1
		ORDER BY s.name
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var show Show
		err := rows.Scan(
			&show.ID, &show.UserID, &show.TVMazeID, &show.Name, &show.Summary, &show.ImageURL, &show.Status, 
			&show.Premiered, &show.Ended, &show.Network, &show.Genres, &show.Rating, &show.CreatedAt, &show.UpdatedAt,
			&show.TotalEpisodes, &show.WatchedEpisodes)
		if err != nil {
			return nil, err
		}
		shows = append(shows, show)
	}
	return shows, nil
}

// MarkAllEpisodesWatched marks all episodes of a show as watched
func MarkAllEpisodesWatched(db *pgxpool.Pool, showID uuid.UUID, userID uuid.UUID) error {
	_, err := db.Exec(context.Background(), `
		INSERT INTO episode_tracking (episode_id, user_id, watched, created_at, updated_at)
		SELECT e.id, $2, true, NOW(), NOW()
		FROM episodes e
		WHERE e.show_id = $1 AND e.user_id = $2
		ON CONFLICT (episode_id, user_id) 
		DO UPDATE SET watched = true, updated_at = NOW()
	`, showID, userID)
	return err
}

func CreateShow(db *pgxpool.Pool, show Show, userID uuid.UUID) (*Show, error) {
	var createdShow Show
	err := db.QueryRow(context.Background(), `
		INSERT INTO shows (user_id, tvmaze_id, name, summary, image_url, status, premiered, ended, network, genres, rating)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, user_id, tvmaze_id, name, summary, image_url, status, premiered, ended, network, genres, rating, created_at, updated_at
	`, userID, show.TVMazeID, show.Name, show.Summary, show.ImageURL, show.Status, show.Premiered, show.Ended, show.Network, show.Genres, show.Rating).Scan(
		&createdShow.ID, &createdShow.UserID, &createdShow.TVMazeID, &createdShow.Name, &createdShow.Summary, 
		&createdShow.ImageURL, &createdShow.Status, &createdShow.Premiered, &createdShow.Ended, &createdShow.Network, 
		&createdShow.Genres, &createdShow.Rating, &createdShow.CreatedAt, &createdShow.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	return &createdShow, nil
}

func GetShowByTVMazeID(db *pgxpool.Pool, tvmazeID int, userID uuid.UUID) (*Show, error) {
	var show Show
	err := db.QueryRow(context.Background(), `
		SELECT id, user_id, tvmaze_id, name, summary, image_url, status, premiered, ended, network, genres, rating, created_at, updated_at 
		FROM shows 
		WHERE tvmaze_id = $1 AND user_id = $2
	`, tvmazeID, userID).Scan(&show.ID, &show.UserID, &show.TVMazeID, &show.Name, &show.Summary, &show.ImageURL, &show.Status, &show.Premiered, &show.Ended, &show.Network, &show.Genres, &show.Rating, &show.CreatedAt, &show.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	return &show, nil
}

func GetShowByID(db *pgxpool.Pool, showID uuid.UUID, userID uuid.UUID) (*Show, error) {
	var show Show
	err := db.QueryRow(context.Background(), `
		SELECT id, user_id, tvmaze_id, name, summary, image_url, status, premiered, ended, network, genres, rating, created_at, updated_at 
		FROM shows 
		WHERE id = $1 AND user_id = $2
	`, showID, userID).Scan(&show.ID, &show.UserID, &show.TVMazeID, &show.Name, &show.Summary, &show.ImageURL, &show.Status, &show.Premiered, &show.Ended, &show.Network, &show.Genres, &show.Rating, &show.CreatedAt, &show.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	return &show, nil
}

func DeleteShow(db *pgxpool.Pool, showID uuid.UUID, userID uuid.UUID) error {
	_, err := db.Exec(context.Background(), `
		DELETE FROM shows WHERE id = $1 AND user_id = $2
	`, showID, userID)
	return err
}

// Episode database functions
func GetEpisodesByShow(db *pgxpool.Pool, showID uuid.UUID, userID uuid.UUID) ([]Episode, error) {
	episodes := []Episode{}

	rows, err := db.Query(context.Background(), `
		SELECT id, show_id, user_id, tvmaze_id, name, season, number, summary, airdate, runtime, image_url, created_at, updated_at 
		FROM episodes 
		WHERE show_id = $1 AND user_id = $2
		ORDER BY season, number
	`, showID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var episode Episode
		err := rows.Scan(&episode.ID, &episode.ShowID, &episode.UserID, &episode.TVMazeID, &episode.Name, &episode.Season, &episode.Number, &episode.Summary, &episode.AirDate, &episode.Runtime, &episode.ImageURL, &episode.CreatedAt, &episode.UpdatedAt)
		if err != nil {
			return nil, err
		}
		episodes = append(episodes, episode)
	}
	return episodes, nil
}

func CreateEpisode(db *pgxpool.Pool, episode Episode) error {
	_, err := db.Exec(context.Background(), `
		INSERT INTO episodes (show_id, user_id, tvmaze_id, name, season, number, summary, airdate, runtime, image_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, episode.ShowID, episode.UserID, episode.TVMazeID, episode.Name, episode.Season, episode.Number, episode.Summary, episode.AirDate, episode.Runtime, episode.ImageURL)
	return err
}

func CreateEpisodes(db *pgxpool.Pool, episodes []Episode) error {
	if len(episodes) == 0 {
		return nil
	}

	// Use simple loop for now instead of batch
	for _, episode := range episodes {
		if err := CreateEpisode(db, episode); err != nil {
			return err
		}
	}
	return nil
}

// Get episodes with tracking data
func GetEpisodesWithTrackingByShow(db *pgxpool.Pool, showID uuid.UUID, userID uuid.UUID) ([]EpisodeWithTracking, error) {
	var episodes []EpisodeWithTracking

	rows, err := db.Query(context.Background(), `
		SELECT 
			e.id, e.show_id, e.user_id, e.tvmaze_id, e.name, e.season, e.number, 
			e.summary, e.airdate, e.runtime, e.image_url, e.created_at, e.updated_at,
			COALESCE(et.watched, false) as watched,
			et.rating, et.notes, et.watched_at
		FROM episodes e
		LEFT JOIN episode_tracking et ON e.id = et.episode_id AND et.user_id = $2
		WHERE e.show_id = $1 AND e.user_id = $2
		ORDER BY e.season DESC, e.number DESC
	`, showID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var episode EpisodeWithTracking
		err := rows.Scan(
			&episode.ID, &episode.ShowID, &episode.UserID, &episode.TVMazeID, &episode.Name, 
			&episode.Season, &episode.Number, &episode.Summary, &episode.AirDate, &episode.Runtime, 
			&episode.ImageURL, &episode.CreatedAt, &episode.UpdatedAt,
			&episode.Watched, &episode.Rating, &episode.Notes, &episode.WatchedAt)
		if err != nil {
			return nil, err
		}
		episodes = append(episodes, episode)
	}
	return episodes, nil
}

// Episode Tracking database functions
func GetEpisodeTracking(db *pgxpool.Pool, episodeID uuid.UUID, userID uuid.UUID) (*EpisodeTracking, error) {
	var tracking EpisodeTracking
	err := db.QueryRow(context.Background(), `
		SELECT id, episode_id, user_id, watched, rating, notes, watched_at, created_at, updated_at 
		FROM episode_tracking 
		WHERE episode_id = $1 AND user_id = $2
	`, episodeID, userID).Scan(&tracking.ID, &tracking.EpisodeID, &tracking.UserID, &tracking.Watched, &tracking.Rating, &tracking.Notes, &tracking.WatchedAt, &tracking.CreatedAt, &tracking.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	return &tracking, nil
}

func CreateOrUpdateEpisodeTracking(db *pgxpool.Pool, tracking EpisodeTracking) error {
	_, err := db.Exec(context.Background(), `
		INSERT INTO episode_tracking (episode_id, user_id, watched, rating, notes, watched_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (episode_id, user_id) 
		DO UPDATE SET 
			watched = EXCLUDED.watched,
			rating = EXCLUDED.rating,
			notes = EXCLUDED.notes,
			watched_at = EXCLUDED.watched_at,
			updated_at = now()
	`, tracking.EpisodeID, tracking.UserID, tracking.Watched, tracking.Rating, tracking.Notes, tracking.WatchedAt)
	return err
}

func MarkEpisodeWatched(db *pgxpool.Pool, episodeID uuid.UUID, userID uuid.UUID, watched bool) error {
	var watchedAt *time.Time
	if watched {
		now := time.Now()
		watchedAt = &now
	}

	_, err := db.Exec(context.Background(), `
		INSERT INTO episode_tracking (episode_id, user_id, watched, watched_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (episode_id, user_id) 
		DO UPDATE SET 
			watched = EXCLUDED.watched,
			watched_at = EXCLUDED.watched_at,
			updated_at = now()
	`, episodeID, userID, watched, watchedAt)
	return err
}

// Get all available seasons for a show
func GetSeasonsByShow(db *pgxpool.Pool, showID uuid.UUID, userID uuid.UUID) ([]int, error) {
	var seasons []int

	rows, err := db.Query(context.Background(), `
		SELECT DISTINCT season 
		FROM episodes 
		WHERE show_id = $1 AND user_id = $2
		ORDER BY season DESC
	`, showID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var season int
		err := rows.Scan(&season)
		if err != nil {
			return nil, err
		}
		seasons = append(seasons, season)
	}
	return seasons, nil
}

// Get episodes with tracking data for a specific season
func GetEpisodesWithTrackingByShowAndSeason(db *pgxpool.Pool, showID uuid.UUID, userID uuid.UUID, season int) ([]EpisodeWithTracking, error) {
	var episodes []EpisodeWithTracking

	rows, err := db.Query(context.Background(), `
		SELECT 
			e.id, e.show_id, e.user_id, e.tvmaze_id, e.name, e.season, e.number, 
			e.summary, e.airdate, e.runtime, e.image_url, e.created_at, e.updated_at,
			COALESCE(et.watched, false) as watched,
			et.rating, et.notes, et.watched_at
		FROM episodes e
		LEFT JOIN episode_tracking et ON e.id = et.episode_id AND et.user_id = $2
		WHERE e.show_id = $1 AND e.user_id = $2 AND e.season = $3
		ORDER BY e.number DESC
	`, showID, userID, season)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var episode EpisodeWithTracking
		err := rows.Scan(
			&episode.ID, &episode.ShowID, &episode.UserID, &episode.TVMazeID, &episode.Name, 
			&episode.Season, &episode.Number, &episode.Summary, &episode.AirDate, &episode.Runtime, 
			&episode.ImageURL, &episode.CreatedAt, &episode.UpdatedAt,
			&episode.Watched, &episode.Rating, &episode.Notes, &episode.WatchedAt)
		if err != nil {
			return nil, err
		}
		episodes = append(episodes, episode)
	}
	return episodes, nil
}