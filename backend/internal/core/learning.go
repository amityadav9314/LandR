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

func (c *LearningCore) AddMaterial(ctx context.Context, userID, matType, content string, existingTags []string) (string, int32, string, []string, error) {
	log.Printf("[Core.AddMaterial] Starting - UserID: %s, Type: %s", userID, matType)

	// 1. Process Content (Scrape if Link)
	finalContent := content
	if matType == "LINK" {
		log.Printf("[Core.AddMaterial] Scraping URL: %s", content)
		scraped, err := c.scraper.Scrape(content)
		if err != nil {
			log.Printf("[Core.AddMaterial] Scraping failed: %v", err)
			return "", 0, "", nil, fmt.Errorf("failed to scrape url: %w", err)
		}
		finalContent = scraped
		log.Printf("[Core.AddMaterial] Scraped content length: %d", len(finalContent))
	}

	// 2. Fetch existing tags if not provided (or merge?)
	// For now, we use what's passed from frontend or fetch from DB if empty?
	// The prompt says "Existing tags you might reuse".
	// Let's fetch all user tags to give context to AI.
	userTags, err := c.store.GetTags(ctx, userID)
	if err != nil {
		log.Printf("[Core.AddMaterial] Failed to fetch tags: %v", err)
		// Continue without tags
	}

	// Combine provided tags and user tags? For now just use userTags for AI context.

	// 3. Generate Flashcards + Title + Tags
	log.Printf("[Core.AddMaterial] Generating flashcards with AI...")
	title, tags, cards, err := c.ai.GenerateFlashcards(finalContent, userTags)
	if err != nil {
		log.Printf("[Core.AddMaterial] AI generation failed: %v", err)
		return "", 0, "", nil, fmt.Errorf("failed to generate flashcards: %w", err)
	}
	log.Printf("[Core.AddMaterial] AI generated Title: %s, Tags: %v, Cards: %d", title, tags, len(cards))

	// 4. Save Material with Title
	log.Printf("[Core.AddMaterial] Saving material to database...")
	materialID, err := c.store.CreateMaterial(ctx, userID, matType, finalContent, title)
	if err != nil {
		log.Printf("[Core.AddMaterial] Failed to save material: %v", err)
		return "", 0, "", nil, fmt.Errorf("failed to create material: %w", err)
	}
	log.Printf("[Core.AddMaterial] Material saved with ID: %s", materialID)

	// 5. Save Tags and Link to Material
	var tagIDs []string
	for _, tagName := range tags {
		tagID, err := c.store.CreateTag(ctx, userID, tagName)
		if err != nil {
			log.Printf("[Core.AddMaterial] Failed to create tag %s: %v", tagName, err)
			continue
		}
		tagIDs = append(tagIDs, tagID)
	}

	if len(tagIDs) > 0 {
		if err := c.store.AddMaterialTags(ctx, materialID, tagIDs); err != nil {
			log.Printf("[Core.AddMaterial] Failed to link tags: %v", err)
		}
	}

	// 6. Save Flashcards
	if len(cards) > 0 {
		log.Printf("[Core.AddMaterial] Saving %d flashcards to database...", len(cards))
		if err := c.store.CreateFlashcards(ctx, materialID, cards); err != nil {
			log.Printf("[Core.AddMaterial] Failed to save flashcards: %v", err)
			return materialID, 0, title, tags, fmt.Errorf("failed to save flashcards: %w", err)
		}
		log.Printf("[Core.AddMaterial] Flashcards saved successfully")
	}

	log.Printf("[Core.AddMaterial] Complete - MaterialID: %s, Cards: %d", materialID, len(cards))
	return materialID, int32(len(cards)), title, tags, nil
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
