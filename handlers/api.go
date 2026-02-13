package handlers

import (
	"context"
	"fmt"
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
	BaseURL string
}

// GetEvent returns full event state with engagement summary
// Route: GET /api/events/{eventIDOrSlug}
func (h *API) GetEvent(w http.ResponseWriter, r *http.Request) {
	eventIDOrSlug := chi.URLParam(r, "eventID")
	ctx := r.Context()

	// Resolve event ID
	eventID, err := h.resolveEventID(ctx, eventIDOrSlug)
	if err != nil {
		h.Log.Printf("Error resolving event '%s': %v", eventIDOrSlug, err)
		writeError(w, http.StatusNotFound, "Event not found")
		return
	}

	// Get event metadata
	event, err := h.Queries.GetEventByID(ctx, eventID)
	if err != nil {
		h.Log.Printf("Error getting event: %v", err)
		writeError(w, http.StatusNotFound, "Event not found")
		return
	}

	// Get all questions for this event
	questions, err := h.Queries.ListQuestionsByEventID(ctx, eventID)
	if err != nil {
		h.Log.Printf("Error getting questions: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// Get total engagement
	totalEngagement, err := h.Queries.GetEventEngagementTotal(ctx, eventID)
	if err != nil {
		h.Log.Printf("Error getting event engagement: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// Get engagement by slug
	bySlugData, err := h.Queries.GetEventEngagementBySlug(ctx, eventID)
	if err != nil {
		h.Log.Printf("Error getting engagement by slug: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// Build questions summary (with vote counts per question)
	questionsSummary := []map[string]interface{}{}
	for i, q := range questions {
		// Get total votes for this question
		qEngagement, err := h.Queries.GetQuestionEngagementTotal(ctx, q.QuestionID)
		if err != nil {
			h.Log.Printf("Error getting question engagement: %v", err)
			continue
		}

		questionsSummary = append(questionsSummary, map[string]interface{}{
			"question_id": q.QuestionID,
			"index":       i + 1,
			"big_text":    q.BigText,
			"sessions":    qEngagement.Sessions,
			"total_votes": qEngagement.TotalVotes,
		})
	}

	// Build response
	response := map[string]interface{}{
		"event_id":        event.EventID,
		"description":     event.Description,
		"created_at":      event.CreatedAt,
		"total_questions": len(questions),
		"engagement": map[string]interface{}{
			"total": map[string]interface{}{
				"sessions":    totalEngagement.Sessions,
				"total_votes": totalEngagement.TotalVotes,
			},
			"by_slug": transformBySlugToMap(bySlugData),
		},
		"questions": questionsSummary,
	}

	if err := writeJSON(w, http.StatusOK, response); err != nil {
		h.Log.Printf("Error writing JSON response: %v", err)
	}
}

// GetQuestions returns all questions with full engagement for an event
// Route: GET /api/events/{eventIDOrSlug}/questions
func (h *API) GetQuestions(w http.ResponseWriter, r *http.Request) {
	eventIDOrSlug := chi.URLParam(r, "eventID")
	ctx := r.Context()

	// Resolve event ID
	eventID, err := h.resolveEventID(ctx, eventIDOrSlug)
	if err != nil {
		h.Log.Printf("Error resolving event '%s': %v", eventIDOrSlug, err)
		writeError(w, http.StatusNotFound, "Event not found")
		return
	}

	// Get all questions
	questions, err := h.Queries.ListQuestionsByEventID(ctx, eventID)
	if err != nil {
		h.Log.Printf("Error getting questions: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// Build response with full engagement for each question
	questionsData := []map[string]interface{}{}
	for i, q := range questions {
		questionData := h.buildQuestionResponse(ctx, q, i+1)
		if questionData != nil {
			questionsData = append(questionsData, questionData)
		}
	}

	response := map[string]interface{}{
		"event_id":  eventID,
		"questions": questionsData,
	}

	if err := writeJSON(w, http.StatusOK, response); err != nil {
		h.Log.Printf("Error writing JSON response: %v", err)
	}
}

// GetQuestion returns a single question with full metadata and engagement
// Route: GET /api/events/{eventIDOrSlug}/questions/{questionIDOrIndex}
// This is the primary endpoint for broadcast graphics (poll every 1-2 seconds)
func (h *API) GetQuestion(w http.ResponseWriter, r *http.Request) {
	eventIDOrSlug := chi.URLParam(r, "eventID")
	questionIDOrIndex := chi.URLParam(r, "questionID")
	ctx := r.Context()

	// Resolve event ID
	eventID, err := h.resolveEventID(ctx, eventIDOrSlug)
	if err != nil {
		h.Log.Printf("Error resolving event '%s': %v", eventIDOrSlug, err)
		writeError(w, http.StatusNotFound, "Event not found")
		return
	}

	// Resolve question
	var question database.Question
	var index int

	if strings.HasPrefix(questionIDOrIndex, "question_") {
		// It's a question ID
		question, err = h.Queries.GetQuestionByID(ctx, questionIDOrIndex)
		if err != nil {
			h.Log.Printf("Error getting question by ID: %v", err)
			writeError(w, http.StatusNotFound, "Question not found")
			return
		}

		// Get the index by counting questions before this one
		questions, _ := h.Queries.ListQuestionsByEventID(ctx, eventID)
		for i, q := range questions {
			if q.QuestionID == questionIDOrIndex {
				index = i + 1
				break
			}
		}
	} else {
		// It's a numeric index
		idx, err := strconv.Atoi(questionIDOrIndex)
		if err != nil {
			h.Log.Printf("Invalid question identifier '%s': %v", questionIDOrIndex, err)
			writeError(w, http.StatusBadRequest, "Invalid question identifier")
			return
		}

		question, err = h.Queries.GetQuestionByEventAndIndex(ctx, database.GetQuestionByEventAndIndexParams{
			EventID:       eventID,
			QuestionIndex: int64(idx),
		})
		if err != nil {
			h.Log.Printf("Error getting question by index %d: %v", idx, err)
			writeError(w, http.StatusNotFound, "Question not found")
			return
		}
		index = idx
	}

	// Build full response
	response := h.buildQuestionResponse(ctx, question, index)
	if response == nil {
		writeError(w, http.StatusInternalServerError, "Error building response")
		return
	}

	if err := writeJSON(w, http.StatusOK, response); err != nil {
		h.Log.Printf("Error writing JSON response: %v", err)
	}
}

// Helper: resolveEventID resolves event ID from slug or returns the ID as-is
func (h *API) resolveEventID(ctx context.Context, eventIDOrSlug string) (string, error) {
	// Check if it's already an event ID
	if strings.HasPrefix(eventIDOrSlug, "event_") {
		return eventIDOrSlug, nil
	}

	// It's a slug - resolve it
	event, err := h.Queries.GetEventBySlug(ctx, eventIDOrSlug)
	if err != nil {
		return "", err
	}

	return event.EventID, nil
}

// Helper: buildQuestionResponse builds a complete question response with engagement
func (h *API) buildQuestionResponse(ctx context.Context, question database.Question, index int) map[string]interface{} {
	// Get engagement totals
	totalEngagement, err := h.Queries.GetQuestionEngagementTotal(ctx, question.QuestionID)
	if err != nil {
		h.Log.Printf("Error getting question engagement total: %v", err)
		return nil
	}

	// Get engagement by slug
	bySlugData, err := h.Queries.GetQuestionEngagementBySlug(ctx, question.QuestionID)
	if err != nil {
		h.Log.Printf("Error getting question engagement by slug: %v", err)
		return nil
	}

	// Calculate percentages
	percentageA, percentageB := calculatePercentages(totalEngagement.VotesA, totalEngagement.VotesB)

	// Build by_slug with percentages
	bySlug := make(map[string]interface{})
	for _, row := range bySlugData {
		pctA, pctB := calculatePercentages(row.VotesA, row.VotesB)
		bySlug[row.Slug] = map[string]interface{}{
			"sessions":     row.Sessions,
			"total_votes":  row.TotalVotes,
			"votes_a":      row.VotesA,
			"votes_b":      row.VotesB,
			"percentage_a": pctA,
			"percentage_b": pctB,
		}
	}

	return map[string]interface{}{
		"question_id": question.QuestionID,
		"event_id":    question.EventID,
		"index":       index,
		"big_text":    question.BigText,
		"small_text":  question.SmallText,
		"image_url":   h.imageURL(question.ImageFilename),
		"choice_a":    question.ChoiceA,
		"choice_b":    question.ChoiceB,
		"engagement": map[string]interface{}{
			"total": map[string]interface{}{
				"sessions":     totalEngagement.Sessions,
				"total_votes":  totalEngagement.TotalVotes,
				"votes_a":      totalEngagement.VotesA,
				"votes_b":      totalEngagement.VotesB,
				"percentage_a": percentageA,
				"percentage_b": percentageB,
			},
			"by_slug": bySlug,
		},
	}
}

// Helper: imageURL constructs full image URL
func (h *API) imageURL(filename string) string {
	return fmt.Sprintf("%s/static/images/%s", h.BaseURL, filename)
}

// Helper: calculatePercentages calculates vote percentages
func calculatePercentages(votesA, votesB interface{}) (float64, float64) {
	// Convert interface{} to int64
	var a, b int64
	if v, ok := votesA.(int64); ok {
		a = v
	}
	if v, ok := votesB.(int64); ok {
		b = v
	}

	total := a + b
	if total == 0 {
		return 0.0, 0.0
	}

	percentageA := float64(a) / float64(total) * 100
	percentageB := float64(b) / float64(total) * 100

	// Round to 2 decimal places
	percentageA = float64(int(percentageA*100)) / 100
	percentageB = float64(int(percentageB*100)) / 100

	return percentageA, percentageB
}
