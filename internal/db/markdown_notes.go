package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateMarkdownNote creates a new markdown note
func CreateMarkdownNote(db *pgxpool.Pool, note *MarkdownNote, userID uuid.UUID) error {
	query := `
		INSERT INTO markdown_notes (id, user_id, title, content, is_rtl, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	noteID := uuid.New()
	now := time.Now()
	
	_, err := DB.Exec(context.Background(), query, noteID, userID, note.Title, note.Content, note.IsRTL, now, now)
	if err != nil {
		return fmt.Errorf("failed to create markdown note: %w", err)
	}
	
	note.ID = noteID
	note.UserID = userID
	note.CreatedAt = now
	note.UpdatedAt = now
	
	return nil
}

// GetMarkdownNoteByID retrieves a markdown note by ID
func GetMarkdownNoteByID(db *pgxpool.Pool, noteID uuid.UUID, userID uuid.UUID) (*MarkdownNote, error) {
	query := `
		SELECT id, user_id, title, content, is_rtl, created_at, updated_at
		FROM markdown_notes
		WHERE id = $1 AND user_id = $2
	`
	
	var note MarkdownNote
	err := DB.QueryRow(context.Background(), query, noteID, userID).Scan(
		&note.ID, &note.UserID, &note.Title, &note.Content, &note.IsRTL,
		&note.CreatedAt, &note.UpdatedAt,
	)
	
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("note not found")
		}
		return nil, fmt.Errorf("failed to get markdown note: %w", err)
	}
	
	return &note, nil
}

// GetAllMarkdownNotes retrieves all markdown notes for a user with pagination
func GetAllMarkdownNotes(db *pgxpool.Pool, userID uuid.UUID, offset, limit int) ([]MarkdownNote, error) {
	query := `
		SELECT id, user_id, title, content, is_rtl, created_at, updated_at
		FROM markdown_notes
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := DB.Query(context.Background(), query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get markdown notes: %w", err)
	}
	defer rows.Close()
	
	var notes []MarkdownNote
	for rows.Next() {
		var note MarkdownNote
		err := rows.Scan(
			&note.ID, &note.UserID, &note.Title, &note.Content, &note.IsRTL,
			&note.CreatedAt, &note.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan markdown note: %w", err)
		}
		notes = append(notes, note)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over markdown notes: %w", err)
	}
	
	return notes, nil
}

// SearchMarkdownNotes searches notes by title and content
func SearchMarkdownNotes(db *pgxpool.Pool, userID uuid.UUID, searchQuery string, offset, limit int) ([]MarkdownNote, error) {
	query := `
		SELECT id, user_id, title, content, is_rtl, created_at, updated_at
		FROM markdown_notes
		WHERE user_id = $1 
		AND (title ILIKE $2 OR content ILIKE $2)
		ORDER BY 
			CASE 
				WHEN title ILIKE $2 THEN 1 
				ELSE 2 
			END,
			updated_at DESC
		LIMIT $3 OFFSET $4
	`
	
	searchPattern := "%" + searchQuery + "%"
	rows, err := DB.Query(context.Background(), query, userID, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search markdown notes: %w", err)
	}
	defer rows.Close()
	
	var notes []MarkdownNote
	for rows.Next() {
		var note MarkdownNote
		err := rows.Scan(
			&note.ID, &note.UserID, &note.Title, &note.Content, &note.IsRTL,
			&note.CreatedAt, &note.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan markdown note: %w", err)
		}
		notes = append(notes, note)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over search results: %w", err)
	}
	
	return notes, nil
}

// UpdateMarkdownNote updates an existing markdown note
func UpdateMarkdownNote(db *pgxpool.Pool, note MarkdownNote, noteID uuid.UUID, userID uuid.UUID) error {
	query := `
		UPDATE markdown_notes
		SET title = $1, content = $2, is_rtl = $3, updated_at = $4
		WHERE id = $5 AND user_id = $6
	`
	
	now := time.Now()
	result, err := DB.Exec(context.Background(), query, note.Title, note.Content, note.IsRTL, now, noteID, userID)
	if err != nil {
		return fmt.Errorf("failed to update markdown note: %w", err)
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("note not found or unauthorized")
	}
	
	return nil
}

// DeleteMarkdownNote deletes a markdown note
func DeleteMarkdownNote(db *pgxpool.Pool, noteID uuid.UUID, userID uuid.UUID) error {
	query := `
		DELETE FROM markdown_notes
		WHERE id = $1 AND user_id = $2
	`
	
	result, err := DB.Exec(context.Background(), query, noteID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete markdown note: %w", err)
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("note not found or unauthorized")
	}
	
	return nil
}

// GetMarkdownNotesCount gets the total count of notes for pagination
func GetMarkdownNotesCount(db *pgxpool.Pool, userID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM markdown_notes
		WHERE user_id = $1
	`
	
	var count int
	err := DB.QueryRow(context.Background(), query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get notes count: %w", err)
	}
	
	return count, nil
}

// GetSearchMarkdownNotesCount gets the count of search results for pagination
func GetSearchMarkdownNotesCount(db *pgxpool.Pool, userID uuid.UUID, searchQuery string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM markdown_notes
		WHERE user_id = $1 
		AND (title ILIKE $2 OR content ILIKE $2)
	`
	
	searchPattern := "%" + searchQuery + "%"
	var count int
	err := DB.QueryRow(context.Background(), query, userID, searchPattern).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get search count: %w", err)
	}
	
	return count, nil
}