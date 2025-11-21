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

func (s *PostgresStore) CreateMaterial(ctx context.Context, userID, matType, content string) (string, error) {
	log.Printf("[Store.CreateMaterial] Inserting material - UserID: %s, Type: %s", userID, matType)
	query := `
		INSERT INTO materials (user_id, type, content)
		VALUES ($1, $2, $3)
		RETURNING id;
	`
	var materialID string
	err := s.db.QueryRow(ctx, query, userID, matType, content).Scan(&materialID)
	if err != nil {
		log.Printf("[Store.CreateMaterial] Insert failed: %v", err)
		return "", fmt.Errorf("failed to insert material: %w", err)
	}
	log.Printf("[Store.CreateMaterial] Material created with ID: %s", materialID)
	return materialID, nil
}

func (s *PostgresStore) CreateFlashcards(ctx context.Context, materialID string, cards []*learning.Flashcard) error {
	log.Printf("[Store.CreateFlashcards] Inserting %d flashcards for material: %s", len(cards), materialID)
	for i, card := range cards {
		query := `
			INSERT INTO flashcards (material_id, question, answer, stage, next_review_at)
			VALUES ($1, $2, $3, $4, $5);
		`
		_, err := s.db.Exec(ctx, query, materialID, card.Question, card.Answer, 0, time.Now().Add(24*time.Hour))
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
		SELECT f.id, f.question, f.answer, f.stage
		FROM flashcards f
		JOIN materials m ON f.material_id = m.id
		WHERE m.user_id = $1 AND f.next_review_at <= NOW()
		ORDER BY f.next_review_at ASC;
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
		if err := rows.Scan(&card.Id, &card.Question, &card.Answer, &card.Stage); err != nil {
			log.Printf("[Store.GetDueFlashcards] Scan failed: %v", err)
			return nil, fmt.Errorf("failed to scan flashcard: %w", err)
		}
		flashcards = append(flashcards, &card)
	}

	log.Printf("[Store.GetDueFlashcards] Found %d flashcards", len(flashcards))
	return flashcards, nil
}

func (s *PostgresStore) UpdateFlashcard(ctx context.Context, id string, stage int32, nextReviewAt time.Time) error {
	log.Printf("[Store.UpdateFlashcard] Updating flashcard: %s, stage: %d", id, stage)
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
