package store

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/amityadav/landr/pkg/pb/auth"
	"github.com/amityadav/landr/pkg/pb/learning"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	db *pgxpool.Pool
}

func NewPostgresStore(ctx context.Context, connString string) (*PostgresStore, error) {
	db, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}
	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}
	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Close() {
	s.db.Close()
}

func (s *PostgresStore) CreateUser(ctx context.Context, email, name, googleID, picture string) (*auth.UserProfile, error) {
	query := `
        INSERT INTO users (email, name, google_id, picture)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (google_id) DO UPDATE
        SET name = EXCLUDED.name, picture = EXCLUDED.picture, updated_at = NOW()
        RETURNING id, email, name, picture;
    `
	row := s.db.QueryRow(ctx, query, email, name, googleID, picture)
	var user auth.UserProfile
	if err := row.Scan(&user.Id, &user.Email, &user.Name, &user.Picture); err != nil {
		return nil, fmt.Errorf("failed to create/update user: %w", err)
	}
	return &user, nil
}

func (s *PostgresStore) GetUserByGoogleID(ctx context.Context, googleID string) (*auth.UserProfile, error) {
	query := `SELECT id, email, name, picture FROM users WHERE google_id = $1`
	row := s.db.QueryRow(ctx, query, googleID)
	var user auth.UserProfile
	if err := row.Scan(&user.Id, &user.Email, &user.Name, &user.Picture); err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (s *PostgresStore) CreateMaterial(ctx context.Context, userID, matType, content, title string) (string, error) {
	log.Printf("[Store.CreateMaterial] Inserting material - UserID: %s, Type: %s, Title: %s", userID, matType, title)
	query := `
        INSERT INTO materials (user_id, type, content, title)
        VALUES ($1, $2, $3, $4)
        RETURNING id;
    `
	var materialID string
	err := s.db.QueryRow(ctx, query, userID, matType, content, title).Scan(&materialID)
	if err != nil {
		log.Printf("[Store.CreateMaterial] Insert failed: %v", err)
		return "", fmt.Errorf("failed to insert material: %w", err)
	}
	log.Printf("[Store.CreateMaterial] Material created with ID: %s", materialID)
	return materialID, nil
}

func (s *PostgresStore) CreateTag(ctx context.Context, userID, name string) (string, error) {
	query := `
		INSERT INTO tags (user_id, name) VALUES ($1, $2)
		ON CONFLICT (user_id, name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id;
	`
	var id string
	err := s.db.QueryRow(ctx, query, userID, name).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to create tag: %w", err)
	}
	return id, nil
}

func (s *PostgresStore) GetTags(ctx context.Context, userID string) ([]string, error) {
	query := `SELECT name FROM tags WHERE user_id = $1 ORDER BY name`
	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tags = append(tags, name)
	}
	return tags, nil
}

func (s *PostgresStore) AddMaterialTags(ctx context.Context, materialID string, tagIDs []string) error {
	for _, tagID := range tagIDs {
		query := `INSERT INTO material_tags (material_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
		if _, err := s.db.Exec(ctx, query, materialID, tagID); err != nil {
			return fmt.Errorf("failed to link tag: %w", err)
		}
	}
	return nil
}

func (s *PostgresStore) GetMaterialTags(ctx context.Context, materialID string) ([]string, error) {
	query := `
		SELECT t.name 
		FROM tags t
		JOIN material_tags mt ON t.id = mt.tag_id
		WHERE mt.material_id = $1
	`
	rows, err := s.db.Query(ctx, query, materialID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tags = append(tags, name)
	}
	return tags, nil
}

func (s *PostgresStore) CreateFlashcards(ctx context.Context, materialID string, cards []*learning.Flashcard) error {
	log.Printf("[Store.CreateFlashcards] Inserting %d flashcards for material: %s", len(cards), materialID)
	for i, card := range cards {
		query := `
            INSERT INTO flashcards (material_id, question, answer, stage, next_review_at)
            VALUES ($1, $2, $3, $4, NOW());
        `
		_, err := s.db.Exec(ctx, query, materialID, card.Question, card.Answer, 0)
		if err != nil {
			log.Printf("[Store.CreateFlashcards] Failed to insert flashcard %d: %v", i, err)
			return fmt.Errorf("failed to insert flashcard: %w", err)
		}
	}
	log.Printf("[Store.CreateFlashcards] All flashcards inserted successfully")
	return nil
}

func (s *PostgresStore) GetDueFlashcards(ctx context.Context, userID string) ([]*learning.Flashcard, error) {
	log.Printf("[Store.GetDueFlashcards] Querying flashcards for userID: %s", userID)
	query := `
        SELECT f.id, f.question, f.answer, f.stage, m.title, m.id
        FROM flashcards f
        JOIN materials m ON f.material_id = m.id
        WHERE m.user_id = $1
        ORDER BY f.id ASC;
    `
	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		log.Printf("[Store.GetDueFlashcards] Query failed: %v", err)
		return nil, fmt.Errorf("failed to query flashcards: %w", err)
	}
	defer rows.Close()

	var flashcards []*learning.Flashcard
	for rows.Next() {
		var card learning.Flashcard
		var title string
		var materialID string
		if err := rows.Scan(&card.Id, &card.Question, &card.Answer, &card.Stage, &title, &materialID); err != nil {
			log.Printf("[Store.GetDueFlashcards] Scan failed: %v", err)
			return nil, fmt.Errorf("failed to scan flashcard: %w", err)
		}
		card.MaterialTitle = title

		tags, err := s.GetMaterialTags(ctx, materialID)
		if err != nil {
			log.Printf("[Store.GetDueFlashcards] Failed to get tags: %v", err)
			// Continue without tags
			tags = []string{}
		}
		card.Tags = tags

		flashcards = append(flashcards, &card)
	}

	return flashcards, nil
}

func (s *PostgresStore) UpdateFlashcard(ctx context.Context, id string, stage int32, nextReviewAt time.Time) error {
	log.Printf("[Store.UpdateFlashcard] Updating flashcard: %s, Stage: %d, NextReviewAt: %v", id, stage, nextReviewAt)
	query := `
        UPDATE flashcards
        SET stage = $1, next_review_at = $2, updated_at = NOW()
        WHERE id = $3;
    `
	_, err := s.db.Exec(ctx, query, stage, nextReviewAt, id)
	if err != nil {
		log.Printf("[Store.UpdateFlashcard] Update failed: %v", err)
		return fmt.Errorf("failed to update flashcard: %w", err)
	}
	log.Printf("[Store.UpdateFlashcard] Flashcard updated successfully")
	return nil
}
