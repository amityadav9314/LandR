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

func (c *LearningCore) GetDueFlashcards(ctx context.Context, userID, materialID string) ([]*learning.Flashcard, error) {
	log.Printf("[Core.GetDueFlashcards] Querying for userID: %s, materialID: %s", userID, materialID)
	cards, err := c.store.GetDueFlashcards(ctx, userID, materialID)
	if err != nil {
		log.Printf("[Core.GetDueFlashcards] Query failed: %v", err)
		return nil, err
	}
	log.Printf("[Core.GetDueFlashcards] Found %d cards", len(cards))
	return cards, nil
}

func (c *LearningCore) GetDueMaterials(ctx context.Context, userID string, page, pageSize int32) ([]*learning.MaterialSummary, int32, error) {
	log.Printf("[Core.GetDueMaterials] Querying for userID: %s, page: %d, pageSize: %d", userID, page, pageSize)
	materials, totalCount, err := c.store.GetDueMaterials(ctx, userID, page, pageSize)
	if err != nil {
		log.Printf("[Core.GetDueMaterials] Query failed: %v", err)
		return nil, 0, err
	}
	log.Printf("[Core.GetDueMaterials] Found %d materials (total: %d)", len(materials), totalCount)
	return materials, totalCount, nil
}

func (c *LearningCore) CompleteReview(ctx context.Context, flashcardID string) error {
	log.Printf("[Core.CompleteReview] Updating flashcard: %s", flashcardID)

	// Fetch the current flashcard to get its stage
	card, err := c.store.GetFlashcard(ctx, flashcardID)
	if err != nil {
		log.Printf("[Core.CompleteReview] Failed to get flashcard: %v", err)
		return fmt.Errorf("failed to get flashcard: %w", err)
	}

	// Implement SRS logic: increment stage and calculate next review time
	// Stage 0: New card -> 1 day
	// Stage 1: 1 day -> 3 days
	// Stage 2: 3 days -> 7 days
	// Stage 3: 7 days -> 15 days
	// Stage 4: 15 days -> 30 days
	// Stage 5+: 30 days (max)

	currentStage := card.Stage
	nextStage := currentStage + 1

	// Calculate next review interval based on new stage
	var intervalDays int
	switch nextStage {
	case 1:
		intervalDays = 1
	case 2:
		intervalDays = 3
	case 3:
		intervalDays = 7
	case 4:
		intervalDays = 15
	default:
		// Stage 5 and above: 30 days
		intervalDays = 30
		if nextStage > 5 {
			nextStage = 5 // Cap at stage 5
		}
	}

	nextReviewAt := time.Now().Add(time.Duration(intervalDays) * 24 * time.Hour)

	log.Printf("[Core.CompleteReview] Advancing from stage %d to %d (next review in %d days)",
		currentStage, nextStage, intervalDays)

	err = c.store.UpdateFlashcard(ctx, flashcardID, nextStage, nextReviewAt)
	if err != nil {
		log.Printf("[Core.CompleteReview] Update failed: %v", err)
		return err
	}

	log.Printf("[Core.CompleteReview] Updated successfully to stage %d", nextStage)
	return nil
}

func (c *LearningCore) GetAllTags(ctx context.Context, userID string) ([]string, error) {
	return c.store.GetTags(ctx, userID)
}

func (c *LearningCore) GetNotificationStatus(ctx context.Context, userID string) (int32, bool, error) {
	log.Printf("[Core.GetNotificationStatus] Getting notification status for userID: %s", userID)

	count, err := c.store.GetDueFlashcardsCount(ctx, userID)
	if err != nil {
		log.Printf("[Core.GetNotificationStatus] Failed to get count: %v", err)
		return 0, false, err
	}

	hasDue := count > 0
	log.Printf("[Core.GetNotificationStatus] User has %d due flashcards", count)
	return count, hasDue, nil
}

func (c *LearningCore) GetMaterialSummary(ctx context.Context, userID, materialID string) (string, string, error) {
	log.Printf("[Core.GetMaterialSummary] Getting summary for materialID: %s, userID: %s", materialID, userID)

	// 1. Fetch material content and existing summary
	content, summary, title, err := c.store.GetMaterialContent(ctx, userID, materialID)
	if err != nil {
		log.Printf("[Core.GetMaterialSummary] Failed to get material: %v", err)
		return "", "", fmt.Errorf("failed to get material: %w", err)
	}

	// 2. If summary exists, return it
	if summary != "" {
		log.Printf("[Core.GetMaterialSummary] Returning existing summary, length: %d", len(summary))
		return summary, title, nil
	}

	// 3. Generate summary via AI
	log.Printf("[Core.GetMaterialSummary] No summary found, generating via AI...")
	summary, err = c.ai.GenerateSummary(content)
	if err != nil {
		log.Printf("[Core.GetMaterialSummary] AI generation failed: %v", err)
		return "", title, fmt.Errorf("failed to generate summary: %w", err)
	}

	// 4. Save summary to database
	if err := c.store.UpdateMaterialSummary(ctx, materialID, summary); err != nil {
		log.Printf("[Core.GetMaterialSummary] Failed to save summary: %v", err)
		// Continue - we can still return the generated summary
	}

	log.Printf("[Core.GetMaterialSummary] Summary generated and saved, length: %d", len(summary))
	return summary, title, nil
}
