package store

import (
	"context"
	"time"

	"github.com/amityadav/landr/backend/proto"
)

type Store interface {
	// User
	CreateUser(ctx context.Context, email, name, googleID, picture string) (*proto.UserProfile, error)
	GetUserByGoogleID(ctx context.Context, googleID string) (*proto.UserProfile, error)

	// Material
	CreateMaterial(ctx context.Context, userID, matType, content string) (string, error)

	// Flashcard
	CreateFlashcards(ctx context.Context, materialID string, cards []*proto.Flashcard) error
	GetDueFlashcards(ctx context.Context, userID string) ([]*proto.Flashcard, error)
	UpdateFlashcard(ctx context.Context, id string, stage int32, nextReviewAt time.Time) error
	
	// General
	Close()
}
