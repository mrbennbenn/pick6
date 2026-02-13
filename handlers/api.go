package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/mrbennbenn/pick6/database"
)

type API struct {
	Queries *database.Queries
	Log     *log.Logger
}

// GetEventEngagement returns event-level engagement metrics
// Route: GET /api/events/{eventIDOrSlug}/engagement
// Accepts either event ID (starts with "event_") or slug
func (h *API) GetEventEngagement(w http.ResponseWriter, r *http.Request) {
	eventIDOrSlug := chi.URLParam(r, "eventID")
	ctx := r.Context()

	// Check if it's an event ID by prefix (avoids unnecessary DB call)
	eventID := eventIDOrSlug
	if !strings.HasPrefix(eventIDOrSlug, "event_") {
		// Not an event ID, must be a slug - resolve it
		event, err := h.Queries.GetEventBySlug(ctx, eventIDOrSlug)
		if err != nil {
			h.Log.Printf("Error resolving event slug '%s': %v", eventIDOrSlug, err)
			writeError(w, http.StatusNotFound, "Event not found")
			return
		}
		eventID = event.EventID
	}

	// Query 1: Get total engagement
	totalData, err := h.Queries.GetEventEngagementTotal(ctx, eventID)
	if err != nil {
		h.Log.Printf("Error getting event engagement total: %v", err)
		writeError(w, http.StatusNotFound, "Event not found")
		return
	}

	// Query 2: Get engagement by slug
	bySlugData, err := h.Queries.GetEventEngagementBySlug(ctx, eventID)
	if err != nil {
		h.Log.Printf("Error getting event engagement by slug: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// Query 3: Get retention by slug
	retentionData, err := h.Queries.GetEventRetentionBySlug(ctx, eventID)
	if err != nil {
		h.Log.Printf("Error getting event retention by slug: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// Build response
	response := map[string]interface{}{
		"event_id": eventID,
		"total": map[string]interface{}{
			"sessions":    totalData.Sessions,
			"total_votes": totalData.TotalVotes,
		},
		"by_slug":           transformBySlugToMap(bySlugData),
		"retention_by_slug": transformRetentionBySlug(retentionData),
	}

	if err := writeJSON(w, http.StatusOK, response); err != nil {
		h.Log.Printf("Error writing JSON response: %v", err)
	}
}

// GetQuestionEngagement returns question-level engagement metrics
// Route: GET /api/events/{eventIDOrSlug}/questions/{questionIDOrIndex}/engagement
// Accepts either event ID (starts with "event_") or slug for event,
// and question ID (starts with "question_") or numeric index (1-based) for question
func (h *API) GetQuestionEngagement(w http.ResponseWriter, r *http.Request) {
	eventIDOrSlug := chi.URLParam(r, "eventID")
	questionIDOrIndex := chi.URLParam(r, "questionID")
	ctx := r.Context()

	// Check if it's an event ID by prefix (avoids unnecessary DB call)
	eventID := eventIDOrSlug
	if !strings.HasPrefix(eventIDOrSlug, "event_") {
		// Not an event ID, must be a slug - resolve it
		event, err := h.Queries.GetEventBySlug(ctx, eventIDOrSlug)
		if err != nil {
			h.Log.Printf("Error resolving event slug '%s': %v", eventIDOrSlug, err)
			writeError(w, http.StatusNotFound, "Event not found")
			return
		}
		eventID = event.EventID
	}

	// Check if it's a question ID by prefix (avoids unnecessary parsing/DB call)
	questionID := questionIDOrIndex
	if !strings.HasPrefix(questionIDOrIndex, "question_") {
		// Not a question ID, must be numeric index - look it up
		index, err := strconv.Atoi(questionIDOrIndex)
		if err != nil {
			h.Log.Printf("Invalid question identifier '%s': not a question ID or numeric index", questionIDOrIndex)
			writeError(w, http.StatusBadRequest, "Invalid question identifier")
			return
		}

		question, err := h.Queries.GetQuestionByEventAndIndex(ctx, database.GetQuestionByEventAndIndexParams{
			EventID:       eventID,
			QuestionIndex: int64(index),
		})
		if err != nil {
			h.Log.Printf("Error getting question by index %d: %v", index, err)
			writeError(w, http.StatusNotFound, "Question not found")
			return
		}
		questionID = question.QuestionID
	}

	// Query 1: Get total engagement for question
	totalData, err := h.Queries.GetQuestionEngagementTotal(ctx, questionID)
	if err != nil {
		h.Log.Printf("Error getting question engagement total: %v", err)
		writeError(w, http.StatusNotFound, "Question not found")
		return
	}

	// Query 2: Get engagement by slug
	bySlugData, err := h.Queries.GetQuestionEngagementBySlug(ctx, questionID)
	if err != nil {
		h.Log.Printf("Error getting question engagement by slug: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// Build response
	response := map[string]interface{}{
		"event_id":    eventID,
		"question_id": questionID,
		"total": map[string]interface{}{
			"sessions":    totalData.Sessions,
			"total_votes": totalData.TotalVotes,
			"votes_a":     totalData.VotesA,
			"votes_b":     totalData.VotesB,
		},
		"by_slug": transformQuestionBySlugToMap(bySlugData),
	}

	if err := writeJSON(w, http.StatusOK, response); err != nil {
		h.Log.Printf("Error writing JSON response: %v", err)
	}
}
