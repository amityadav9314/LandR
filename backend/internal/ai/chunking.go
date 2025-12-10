package ai

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/amityadav/landr/pkg/pb/learning"
)

const (
	ChunkSize      = 3000 // ~750 tokens
	ChunkOverlap   = 200  // Overlap between chunks
	MaxRetries     = 3
	BaseRetryDelay = 2 * time.Second
)

// ChunkResult holds the result from processing a single chunk
type ChunkResult struct {
	Title      string
	Tags       []string
	Flashcards []*learning.Flashcard
	Error      error
	ChunkIndex int
}

// EstimateTokens estimates token count (roughly chars/4)
func EstimateTokens(text string) int {
	return len(text) / 4
}

// SplitIntoChunks splits text into overlapping chunks
func SplitIntoChunks(text string, chunkSize, overlap int) []string {
	if len(text) <= chunkSize {
		return []string{text}
	}

	var chunks []string
	start := 0

	for start < len(text) {
		end := start + chunkSize
		if end > len(text) {
			end = len(text)
		}

		// Try to break at sentence boundary
		if end < len(text) {
			// Look for sentence end in last 100 chars
			searchStart := end - 100
			if searchStart < start {
				searchStart = start
			}
			chunk := text[searchStart:end]

			// Find last sentence boundary
			lastPeriod := strings.LastIndex(chunk, ". ")
			if lastPeriod > 0 {
				end = searchStart + lastPeriod + 2
			}
		}

		chunks = append(chunks, strings.TrimSpace(text[start:end]))
		start = end - overlap
	}

	log.Printf("[Chunker] Split text into %d chunks", len(chunks))
	return chunks
}

// RetryWithBackoff retries a function with exponential backoff
func RetryWithBackoff(ctx context.Context, operation string, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < MaxRetries; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err

		// Check if retryable error
		errStr := err.Error()
		isRetryable := strings.Contains(errStr, "429") ||
			strings.Contains(errStr, "500") ||
			strings.Contains(errStr, "502") ||
			strings.Contains(errStr, "503") ||
			strings.Contains(errStr, "timeout") ||
			strings.Contains(errStr, "connection refused")

		if !isRetryable {
			log.Printf("[Retry.%s] Non-retryable error: %v", operation, err)
			return err
		}

		// Calculate delay with exponential backoff + jitter
		delay := time.Duration(math.Pow(2, float64(attempt))) * BaseRetryDelay
		jitter := time.Duration(rand.Int63n(int64(delay / 2)))
		delay += jitter

		log.Printf("[Retry.%s] Attempt %d failed: %v. Retrying in %v...", operation, attempt+1, err, delay)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// ProcessChunksParallel processes multiple chunks in parallel using goroutines
func (c *Client) ProcessChunksParallel(ctx context.Context, chunks []string, existingTags []string) (string, []string, []*learning.Flashcard, error) {
	if len(chunks) == 0 {
		return "", nil, nil, fmt.Errorf("no chunks to process")
	}

	if len(chunks) == 1 {
		// Single chunk - process normally
		return c.GenerateFlashcards(chunks[0], existingTags)
	}

	log.Printf("[AI.Parallel] Processing %d chunks in parallel", len(chunks))

	// Channel to collect results
	resultsChan := make(chan ChunkResult, len(chunks))

	// Use semaphore to limit concurrent requests (avoid rate limits)
	maxConcurrent := 3
	sem := make(chan struct{}, maxConcurrent)

	var wg sync.WaitGroup

	// Process each chunk in a goroutine
	for i, chunk := range chunks {
		wg.Add(1)
		go func(idx int, chunkText string) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			result := ChunkResult{ChunkIndex: idx}

			// Retry with backoff
			err := RetryWithBackoff(ctx, fmt.Sprintf("Chunk_%d", idx), func() error {
				title, tags, cards, err := c.GenerateFlashcards(chunkText, existingTags)
				if err != nil {
					return err
				}
				result.Title = title
				result.Tags = tags
				result.Flashcards = cards
				return nil
			})

			if err != nil {
				result.Error = err
			}

			resultsChan <- result
		}(i, chunk)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	var allFlashcards []*learning.Flashcard
	allTags := make(map[string]bool)
	var firstTitle string
	var errors []error

	for result := range resultsChan {
		if result.Error != nil {
			errors = append(errors, result.Error)
			continue
		}

		if firstTitle == "" && result.Title != "" {
			firstTitle = result.Title
		}

		for _, tag := range result.Tags {
			allTags[tag] = true
		}

		allFlashcards = append(allFlashcards, result.Flashcards...)
	}

	// If all chunks failed, return error
	if len(errors) == len(chunks) {
		return "", nil, nil, fmt.Errorf("all chunks failed: %v", errors[0])
	}

	// Convert tags map to slice
	var tags []string
	for tag := range allTags {
		tags = append(tags, tag)
	}

	// Deduplicate flashcards
	dedupedCards := deduplicateFlashcards(allFlashcards)

	log.Printf("[AI.Parallel] Completed: %d unique flashcards from %d chunks", len(dedupedCards), len(chunks))
	return firstTitle, tags, dedupedCards, nil
}

// deduplicateFlashcards removes duplicate flashcards based on question similarity
func deduplicateFlashcards(cards []*learning.Flashcard) []*learning.Flashcard {
	seen := make(map[string]bool)
	var unique []*learning.Flashcard

	for _, card := range cards {
		// Normalize question for comparison
		key := strings.ToLower(strings.TrimSpace(card.Question))
		if len(key) > 50 {
			key = key[:50] // Use first 50 chars as key
		}

		if !seen[key] {
			seen[key] = true
			unique = append(unique, card)
		}
	}

	return unique
}
