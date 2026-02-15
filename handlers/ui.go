package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/mrbennbenn/pick6/database"
	"github.com/mrbennbenn/pick6/middleware"
	"github.com/mrbennbenn/pick6/templates"
)

type UI struct {
	Queries *database.Queries
	Log     *log.Logger
}

// RedirectToFirst redirects to the first question
// Route: GET /{slug}/
func (h *UI) RedirectToFirst(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	// Validate slug exists
	_, err := h.Queries.GetEventBySlug(r.Context(), slug)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		h.Log.Printf("Error getting event by slug: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Redirect to question 1
	http.Redirect(w, r, fmt.Sprintf("/%s/question/1", slug), http.StatusSeeOther)
}

// ShowQuestion displays a question form with progress and existing votes
// Route: GET /{slug}/question/{order}
func (h *UI) ShowQuestion(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	orderStr := chi.URLParam(r, "order")

	// Parse order
	order, err := strconv.Atoi(orderStr)
	if err != nil || order < 1 {
		http.NotFound(w, r)
		return
	}
	currentIndex := order - 1 // Convert to 0-based

	// Get session ID
	sessionID, err := middleware.SessionFromCtx(r.Context())
	if err != nil {
		h.Log.Printf("Error getting session from context: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get event (with retry logic for transient failures)
	var event database.Event
	err = database.WithRetry(r.Context(), database.DefaultRetryConfig(), func() error {
		var queryErr error
		event, queryErr = h.Queries.GetEventBySlug(r.Context(), slug)
		return queryErr
	})
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		h.Log.Printf("Error getting event: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get all questions (needed for progress indicator in template) with retry
	var questions []database.Question
	err = database.WithRetry(r.Context(), database.DefaultRetryConfig(), func() error {
		var queryErr error
		questions, queryErr = h.Queries.ListQuestionsByEventID(r.Context(), event.EventID)
		return queryErr
	})
	if err != nil {
		h.Log.Printf("Error listing questions: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Validate index
	if currentIndex < 0 || currentIndex >= len(questions) {
		http.NotFound(w, r)
		return
	}

	// Get existing responses (with retry)
	var responses []database.Response
	err = database.WithRetry(r.Context(), database.DefaultRetryConfig(), func() error {
		var queryErr error
		responses, queryErr = h.Queries.GetResponsesBySessionAndEvent(r.Context(),
			database.GetResponsesBySessionAndEventParams{
				SessionID: sessionID,
				EventID:   event.EventID,
			})
		return queryErr
	})
	if err != nil {
		h.Log.Printf("Error getting responses: %v", err)
		// Continue without existing responses
		responses = []database.Response{}
	}

	// Build view model
	vm := templates.QuestionViewModel{
		Slug:            slug,
		Questions:       convertToTemplateQuestions(questions),
		CurrentIndex:    currentIndex,
		ExistingAnswers: buildExistingAnswersMap(responses),
		Errors:          parseErrors(r),
	}

	// Prevent browser caching to ensure fresh data on back button
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Render template
	if err := templates.QuestionPage(vm).Render(r.Context(), w); err != nil {
		h.Log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// SubmitAnswer saves user's answer and redirects to next question or submit-info
// Route: POST /{slug}/question/{order}
func (h *UI) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	orderStr := chi.URLParam(r, "order")

	// Parse order
	order, err := strconv.Atoi(orderStr)
	if err != nil || order < 1 {
		http.NotFound(w, r)
		return
	}

	// Get session ID
	sessionID, err := middleware.SessionFromCtx(r.Context())
	if err != nil {
		h.Log.Printf("Error getting session from context: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get event (with retry logic)
	var event database.Event
	err = database.WithRetry(r.Context(), database.DefaultRetryConfig(), func() error {
		var queryErr error
		event, queryErr = h.Queries.GetEventBySlug(r.Context(), slug)
		return queryErr
	})
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		h.Log.Printf("Error getting event: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Validate index range (questions are 1-6, so order must be 1-6)
	if order < 1 || order > 6 {
		http.NotFound(w, r)
		return
	}

	// Fetch only the current question instead of all questions (optimization) with retry
	var currentQuestion database.Question
	err = database.WithRetry(r.Context(), database.DefaultRetryConfig(), func() error {
		var queryErr error
		currentQuestion, queryErr = h.Queries.GetQuestionByEventAndIndex(r.Context(), database.GetQuestionByEventAndIndexParams{
			EventID:       event.EventID,
			QuestionIndex: order,
		})
		return queryErr
	})
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		h.Log.Printf("Error getting question: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	choice := r.FormValue("choice")

	// Validate choice
	if choice != "a" && choice != "b" {
		redirectURL := buildErrorRedirectURL(
			fmt.Sprintf("/%s/question/%d", slug, order),
			map[string]string{"choice": "Please select a fighter"},
			nil,
		)
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}

	// Save response (with retry logic)
	err = database.WithRetry(r.Context(), database.DefaultRetryConfig(), func() error {
		_, queryErr := h.Queries.UpsertResponse(r.Context(), database.UpsertResponseParams{
			QuestionID: currentQuestion.QuestionID,
			SessionID:  sessionID,
			Slug:       slug,
			Choice:     choice,
		})
		return queryErr
	})
	if err != nil {
		h.Log.Printf("Error saving response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Determine next step (hardcoded to 6 questions)
	if order == 6 {
		// Last question - redirect to info form
		http.Redirect(w, r, fmt.Sprintf("/%s/submit-info", slug), http.StatusSeeOther)
	} else {
		// More questions - redirect to next
		nextOrder := order + 1
		http.Redirect(w, r, fmt.Sprintf("/%s/question/%d", slug, nextOrder), http.StatusSeeOther)
	}
}

// ShowInfoForm displays the user information collection form
// Route: GET /{slug}/submit-info
func (h *UI) ShowInfoForm(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	// Validate slug exists
	_, err := h.Queries.GetEventBySlug(r.Context(), slug)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		h.Log.Printf("Error getting event: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Build view model (pre-fill from query params if validation failed)
	vm := templates.InfoFormViewModel{
		Slug:   slug,
		Name:   r.URL.Query().Get("name"),
		Email:  r.URL.Query().Get("email"),
		Phone:  r.URL.Query().Get("phone"),
		Errors: parseErrors(r),
	}

	// Render template
	if err := templates.InfoFormPage(vm).Render(r.Context(), w); err != nil {
		h.Log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// SubmitInfoForm saves user information and redirects to end page
// Route: POST /{slug}/submit-info
func (h *UI) SubmitInfoForm(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	// Get session ID
	sessionID, err := middleware.SessionFromCtx(r.Context())
	if err != nil {
		h.Log.Printf("Error getting session from context: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))
	phone := strings.TrimSpace(r.FormValue("phone"))

	// Validate
	errors := make(map[string]string)

	if name == "" {
		errors["name"] = "Name is required"
	}

	if email == "" {
		errors["email"] = "Email is required"
	} else if !isValidEmail(email) {
		errors["email"] = "Please enter a valid email address"
	}

	if phone == "" {
		errors["phone"] = "Phone number is required"
	} else {
		valid, normalizedPhone, errorMsg := isValidPhone(phone, "GB")
		if !valid {
			errors["phone"] = errorMsg
		} else {
			// Use normalized phone for storage
			phone = normalizedPhone
		}
	}

	// If validation fails, redirect back with errors
	if len(errors) > 0 {
		redirectURL := buildErrorRedirectURL(
			fmt.Sprintf("/%s/submit-info", slug),
			errors,
			map[string]string{
				"name":  name,
				"email": email,
				"phone": r.FormValue("phone"), // Use original, not normalized
			},
		)
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}

	// Save session
	_, err = h.Queries.UpsertSession(r.Context(), database.UpsertSessionParams{
		SessionID: sessionID,
		Name:      sql.NullString{String: name, Valid: true},
		Email:     sql.NullString{String: email, Valid: true},
		Mobile:    sql.NullString{String: phone, Valid: true}, // E.164 format
	})
	if err != nil {
		h.Log.Printf("Error saving session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Redirect to end page
	http.Redirect(w, r, fmt.Sprintf("/%s/end", slug), http.StatusSeeOther)
}

// ShowEnd displays the thank you page
// Route: GET /{slug}/end
func (h *UI) ShowEnd(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	// Get session ID
	sessionID, err := middleware.SessionFromCtx(r.Context())
	if err != nil {
		h.Log.Printf("Error getting session from context: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get event
	event, err := h.Queries.GetEventBySlug(r.Context(), slug)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		h.Log.Printf("Error getting event: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get user's responses to count total answers
	responses, err := h.Queries.GetResponsesBySessionAndEvent(r.Context(),
		database.GetResponsesBySessionAndEventParams{
			SessionID: sessionID,
			EventID:   event.EventID,
		})
	if err != nil {
		h.Log.Printf("Error getting responses: %v", err)
		// Continue with zero count
		responses = []database.Response{}
	}

	// Build view model
	vm := templates.EndViewModel{
		Slug:         slug,
		TotalAnswers: len(responses),
	}

	// Render template
	if err := templates.EndPage(vm).Render(r.Context(), w); err != nil {
		h.Log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
