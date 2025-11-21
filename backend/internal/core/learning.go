package core

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/amityadav/landr/internal/ai"
	"github.com/amityadav/landr/internal/scraper"
	"github.com/amityadav/landr/internal/store"
	"github.com/amityadav/landr/pkg/pb/learning"
)

type LearningCore struct {
	store   store.Store
	scraper *scraper.Scraper
	ai      *ai.Client
}

func NewLearningCore(s store.Store, scraper *scraper.Scraper, ai *ai.Client) *LearningCore {
	return &LearningCore{
		store:   s,
		scraper: scraper,
		ai:      ai,
	}
}

func (c *LearningCore) AddMaterial(ctx context.Context, userID, matType, content string) (string, int32, error) {
	log.Printf("[Core.AddMaterial] Starting - UserID: %s, Type: %s", userID, matType)

	// 1. Process Content (Scrape if Link)
	finalContent := content
	if matType == "LINK" {
		log.Printf("[Core.AddMaterial] Scraping URL: %s", content)
		scraped, err := c.scraper.Scrape(content)
		if err != nil {
			log.Printf("[Core.AddMaterial] Scraping failed: %v", err)
			return "", 0, fmt.Errorf("failed to scrape url: %w", err)
		}
		finalContent = scraped
		log.Printf("[Core.AddMaterial] Scraped content length: %d", len(finalContent))
	}

	// 2. Save Material
	log.Printf("[Core.AddMaterial] Saving material to database...")
	materialID, err := c.store.CreateMaterial(ctx, userID, matType, finalContent)
	if err != nil {
		log.Printf("[Core.AddMaterial] Failed to save material: %v", err)
		return "", 0, fmt.Errorf("failed to create material: %w", err)
	}
	log.Printf("[Core.AddMaterial] Material saved with ID: %s", materialID)

	// 3. Generate Flashcards
	log.Printf("[Core.AddMaterial] Generating flashcards with AI...")
	cards, err := c.ai.GenerateFlashcards(finalContent)
	if err != nil {
		log.Printf("[Core.AddMaterial] AI generation failed: %v", err)
		// Note: We might want to handle this gracefully (e.g., retry later),
		// but for now we fail the request or return 0 cards to avoid losing the material.
		// Let's log it and return success with 0 cards to avoid losing the material.
		// For this MVP, returning error is safer to let user know AI failed.
		return materialID, 0, fmt.Errorf("failed to generate flashcards: %w", err)
	}
	log.Printf("[Core.AddMaterial] AI generated %d flashcards", len(cards))

	// 4. Save Flashcards
	if len(cards) > 0 {
		log.Printf("[Core.AddMaterial] Saving %d flashcards to database...", len(cards))
		if err := c.store.CreateFlashcards(ctx, materialID, cards); err != nil {
			log.Printf("[Core.AddMaterial] Failed to save flashcards: %v", err)
			return materialID, 0, fmt.Errorf("failed to save flashcards: %w", err)
		}
		log.Printf("[Core.AddMaterial] Flashcards saved successfully")
	}

	log.Printf("[Core.AddMaterial] Complete - MaterialID: %s, Cards: %d", materialID, len(cards))
	return materialID, int32(len(cards)), nil
}

func (c *LearningCore) GetDueFlashcards(ctx context.Context, userID string) ([]*learning.Flashcard, error) {
	log.Printf("[Core.GetDueFlashcards] Querying for userID: %s", userID)
	cards, err := c.store.GetDueFlashcards(ctx, userID)
	if err != nil {
		log.Printf("[Core.GetDueFlashcards] Query failed: %v", err)
		return nil, err
	}
	log.Printf("[Core.GetDueFlashcards] Found %d cards", len(cards))
	return cards, nil
}

func (c *LearningCore) CompleteReview(ctx context.Context, flashcardID string) error {
	log.Printf("[Core.CompleteReview] Updating flashcard: %s", flashcardID)
	// Simple SRS logic: Always bump to next stage for now, or just reset.
	// For MVP:
	// Stage 1: 1 day
	// Stage 2: 3 days
	// Stage 3: 7 days
	// Stage 4: 15 days
	// Stage 5: 30 days

	// We need to fetch the card first to know current stage.
	// For now, let's assume we just increment stage and set review time.
	// This requires a GetFlashcard method or we pass current stage from frontend.
	// Let's implement a simple update: UpdateFlashcard(id, stage+1, now + interval)

	// Since we don't have the current stage in the request, we'll assume the user sent the review
	// and we just want to mark it reviewed.
	// Ideally, we should read the card. Let's keep it simple:
	// We will just update the review time to +1 day for now (Stage 1).
	// TODO: Implement full SRS logic by reading card first.

	err := c.store.UpdateFlashcard(ctx, flashcardID, 1, time.Now().Add(24*time.Hour))
	if err != nil {
		log.Printf("[Core.CompleteReview] Update failed: %v", err)
		return err
	}
	log.Printf("[Core.CompleteReview] Updated successfully")
	return nil
}
