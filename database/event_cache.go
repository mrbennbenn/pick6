package database

import (
	"context"
	"fmt"
	"time"

	cache "github.com/patrickmn/go-cache"
)

// EventCache caches event and questions data to reduce database load
// Events and questions are essentially static during an event's lifecycle
type EventCache struct {
	cache   *cache.Cache
	queries *Queries
}

// CachedEventData contains event and its questions
type CachedEventData struct {
	Event     Event
	Questions []Question
}

// NewEventCache creates a new event cache
// defaultTTL: how long to cache (recommend 1 hour for static event data)
// cleanupInterval: how often to cleanup expired entries
func NewEventCache(queries *Queries, defaultTTL, cleanupInterval time.Duration) *EventCache {
	return &EventCache{
		cache:   cache.New(defaultTTL, cleanupInterval),
		queries: queries,
	}
}

// GetEventWithQuestionsBySlug retrieves event and questions from cache or database
// Cache key is the slug, value contains both event and questions
func (ec *EventCache) GetEventWithQuestionsBySlug(ctx context.Context, slug string) (*CachedEventData, error) {
	// Check cache first
	if cached, found := ec.cache.Get(slug); found {
		if data, ok := cached.(*CachedEventData); ok {
			return data, nil
		}
	}

	// Cache miss - fetch from database
	event, err := ec.queries.GetEventBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get event by slug: %w", err)
	}

	questions, err := ec.queries.ListQuestionsByEventID(ctx, event.EventID)
	if err != nil {
		return nil, fmt.Errorf("failed to list questions: %w", err)
	}

	// Store in cache
	data := &CachedEventData{
		Event:     event,
		Questions: questions,
	}
	ec.cache.Set(slug, data, cache.DefaultExpiration)

	return data, nil
}

// InvalidateSlug removes a slug from the cache
// Useful if event data changes (rare, but possible)
func (ec *EventCache) InvalidateSlug(slug string) {
	ec.cache.Delete(slug)
}

// InvalidateAll clears the entire cache
func (ec *EventCache) InvalidateAll() {
	ec.cache.Flush()
}

// Stats returns cache statistics for monitoring
func (ec *EventCache) Stats() (items int, hits, misses uint64) {
	return ec.cache.ItemCount(), 0, 0 // go-cache doesn't track hits/misses
}
