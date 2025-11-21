package service

import (
	"context"
	"log"

	"github.com/amityadav/landr/internal/core"
	"github.com/amityadav/landr/pkg/pb/learning"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type LearningService struct {
	learning.UnimplementedLearningServiceServer
	core *core.LearningCore
}

func NewLearningService(c *core.LearningCore) *LearningService {
	return &LearningService{
		core: c,
	}
}

func (s *LearningService) AddMaterial(ctx context.Context, req *learning.AddMaterialRequest) (*learning.AddMaterialResponse, error) {
	log.Printf("[AddMaterial] Received request - Type: %s, Content length: %d", req.Type, len(req.Content))

	// We need the userID. In a real app, this comes from the context (via interceptor).
	// For this MVP, we'll assume a hardcoded user or extract from metadata if we implemented auth interceptor.
	// Let's assume a placeholder userID for now since we haven't built the interceptor yet.
	// TODO: Extract userID from context.
	userID := "8bd27697-d4a9-405e-b0ca-e0ee78ef045b" // Test user UUID
	log.Printf("[AddMaterial] Using userID: %s", userID)

	materialID, count, err := s.core.AddMaterial(ctx, userID, req.Type, req.Content)
	if err != nil {
		log.Printf("[AddMaterial] ERROR: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to add material: %v", err)
	}

	log.Printf("[AddMaterial] SUCCESS - MaterialID: %s, Flashcards created: %d", materialID, count)
	return &learning.AddMaterialResponse{
		MaterialId:        materialID,
		FlashcardsCreated: count,
	}, nil
}

func (s *LearningService) GetDueFlashcards(ctx context.Context, _ *emptypb.Empty) (*learning.FlashcardList, error) {
	userID := "8bd27697-d4a9-405e-b0ca-e0ee78ef045b" // Test user UUID
	log.Printf("[GetDueFlashcards] Fetching flashcards for userID: %s", userID)

	cards, err := s.core.GetDueFlashcards(ctx, userID)
	if err != nil {
		log.Printf("[GetDueFlashcards] ERROR: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get due flashcards: %v", err)
	}

	log.Printf("[GetDueFlashcards] SUCCESS - Found %d flashcards", len(cards))
	return &learning.FlashcardList{
		Flashcards: cards,
	}, nil
}

func (s *LearningService) CompleteReview(ctx context.Context, req *learning.CompleteReviewRequest) (*emptypb.Empty, error) {
	log.Printf("[CompleteReview] Completing review for flashcardID: %s", req.FlashcardId)

	if err := s.core.CompleteReview(ctx, req.FlashcardId); err != nil {
		log.Printf("[CompleteReview] ERROR: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to complete review: %v", err)
	}

	log.Printf("[CompleteReview] SUCCESS")
	return &emptypb.Empty{}, nil
}
