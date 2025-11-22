package store

import (
	"context"
	"time"

	"github.com/amityadav/landr/pkg/pb/auth"
	"github.com/amityadav/landr/pkg/pb/learning"
)

type Store interface {
	// User
	CreateUser(ctx context.Context, email, name, googleID, picture string) (*auth.UserProfile, error)
	GetUserByGoogleID(ctx context.Context, googleID string) (*auth.UserProfile, error)

	// Material
	CreateMaterial(ctx context.Context, userID, matType, content, title string) (string, error)

	// Tags
	CreateTag(ctx context.Context, userID, name string) (string, error)
	GetTags(ctx context.Context, userID string) ([]string, error)
	AddMaterialTags(ctx context.Context, materialID string, tagIDs []string) error
	GetMaterialTags(ctx context.Context, materialID string) ([]string, error)

	// Flashcard
	CreateFlashcards(ctx context.Context, materialID string, cards []*learning.Flashcard) error
	GetDueFlashcards(ctx context.Context, userID, materialID string) ([]*learning.Flashcard, error)
	GetDueMaterials(ctx context.Context, userID string) ([]*learning.MaterialSummary, error)
	UpdateFlashcard(ctx context.Context, id string, stage int32, nextReviewAt time.Time) error

	// General
	Close()
}
