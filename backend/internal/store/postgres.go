package store

import (
	"context"
	"fmt"
	"time"

	"github.com/amityadav/landr/backend/proto"
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

// Placeholder implementations to satisfy interface (will implement fully in next steps)
func (s *PostgresStore) CreateUser(ctx context.Context, email, name, googleID, picture string) (*proto.UserProfile, error) {
	query := `
		INSERT INTO users (email, name, google_id, picture)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (google_id) DO UPDATE 
		SET name = EXCLUDED.name, picture = EXCLUDED.picture, updated_at = NOW()
		RETURNING id, email, name, picture;
	`
	row := s.db.QueryRow(ctx, query, email, name, googleID, picture)

	var user proto.UserProfile
	if err := row.Scan(&user.Id, &user.Email, &user.Name, &user.Picture); err != nil {
		return nil, fmt.Errorf("failed to create/update user: %w", err)
	}
	return &user, nil
}

func (s *PostgresStore) GetUserByGoogleID(ctx context.Context, googleID string) (*proto.UserProfile, error) {
	query := `SELECT id, email, name, picture FROM users WHERE google_id = $1`
	row := s.db.QueryRow(ctx, query, googleID)

	var user proto.UserProfile
	if err := row.Scan(&user.Id, &user.Email, &user.Name, &user.Picture); err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}
func (s *PostgresStore) CreateMaterial(ctx context.Context, userID, matType, content string) (string, error) {
	return "", nil
}
func (s *PostgresStore) CreateFlashcards(ctx context.Context, materialID string, cards []*proto.Flashcard) error {
	return nil
}
func (s *PostgresStore) GetDueFlashcards(ctx context.Context, userID string) ([]*proto.Flashcard, error) {
	return nil, nil
}
func (s *PostgresStore) UpdateFlashcard(ctx context.Context, id string, stage int32, nextReviewAt time.Time) error {
	return nil
}
