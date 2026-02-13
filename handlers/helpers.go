package handlers

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"net/url"
	"strings"

	"github.com/mrbennbenn/pick6/database"
	"github.com/mrbennbenn/pick6/templates"
	"github.com/nyaruka/phonenumbers"
)

// parseErrors extracts error messages from query parameters
// Looks for parameters like error_name, error_email, etc.
func parseErrors(r *http.Request) map[string]string {
	errors := make(map[string]string)
	for key, values := range r.URL.Query() {
		if len(values) > 0 && strings.HasPrefix(key, "error_") {
			field := strings.TrimPrefix(key, "error_")
			errors[field] = values[0]
		}
	}
	return errors
}

// buildExistingAnswersMap converts database responses to a map[questionID]choice
func buildExistingAnswersMap(responses []database.Response) map[string]string {
	answers := make(map[string]string)
	for _, resp := range responses {
		answers[resp.QuestionID] = resp.Choice
	}
	return answers
}

// convertToTemplateQuestions converts database.Question to templates.Question
func convertToTemplateQuestions(dbQuestions []database.Question) []templates.Question {
	questions := make([]templates.Question, len(dbQuestions))
	for i, q := range dbQuestions {
		questions[i] = templates.Question{
			QuestionID:    q.QuestionID,
			BigText:       q.BigText,
			SmallText:     q.SmallText,
			ImageFilename: q.ImageFilename,
			ChoiceA:       q.ChoiceA,
			ChoiceB:       q.ChoiceB,
		}
	}
	return questions
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// writeError writes a plain text error response
func writeError(w http.ResponseWriter, status int, message string) {
	http.Error(w, message, status)
}

// isValidEmail validates email format using net/mail
func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// isValidPhone validates and normalizes phone number
// Returns: (isValid, normalizedPhone, errorMessage)
// defaultRegion should be "GB" for UK-based numbers
// Only accepts mobile numbers, returns E.164 format
func isValidPhone(phone, defaultRegion string) (bool, string, string) {
	// Parse phone number with default region
	num, err := phonenumbers.Parse(phone, defaultRegion)
	if err != nil {
		return false, "", "Invalid phone number format"
	}

	// Validate number
	if !phonenumbers.IsValidNumber(num) {
		return false, "", "Please enter a valid phone number"
	}

	// Check if mobile only
	numberType := phonenumbers.GetNumberType(num)
	if numberType != phonenumbers.MOBILE && numberType != phonenumbers.FIXED_LINE_OR_MOBILE {
		return false, "", "Please enter a mobile number"
	}

	// Format to E.164 for storage
	normalized := phonenumbers.Format(num, phonenumbers.E164)

	return true, normalized, ""
}

// transformBySlugToMap converts event engagement slice to map[slug]data
func transformBySlugToMap(rows []database.GetEventEngagementBySlugRow) map[string]interface{} {
	result := make(map[string]interface{})
	for _, row := range rows {
		result[row.Slug] = map[string]interface{}{
			"sessions":    row.Sessions,
			"total_votes": row.TotalVotes,
		}
	}
	return result
}

// transformRetentionBySlug groups retention data by slug
func transformRetentionBySlug(rows []database.GetEventRetentionBySlugRow) map[string][]interface{} {
	result := make(map[string][]interface{})
	for _, row := range rows {
		slug := row.Slug
		if _, exists := result[slug]; !exists {
			result[slug] = []interface{}{}
		}
		result[slug] = append(result[slug], map[string]interface{}{
			"question_id":       row.QuestionID,
			"big_text":          row.BigText,
			"sessions_answered": row.SessionsAnswered,
		})
	}
	return result
}

// transformQuestionBySlugToMap converts question engagement to map
func transformQuestionBySlugToMap(rows []database.GetQuestionEngagementBySlugRow) map[string]interface{} {
	result := make(map[string]interface{})
	for _, row := range rows {
		result[row.Slug] = map[string]interface{}{
			"sessions":    row.Sessions,
			"total_votes": row.TotalVotes,
			"votes_a":     row.VotesA,
			"votes_b":     row.VotesB,
		}
	}
	return result
}

// buildErrorRedirectURL builds a redirect URL with error and pre-fill query parameters
func buildErrorRedirectURL(baseURL string, errors map[string]string, values map[string]string) string {
	queryParams := url.Values{}

	// Add pre-fill values
	for key, value := range values {
		if value != "" {
			queryParams.Set(key, value)
		}
	}

	// Add errors
	for field, msg := range errors {
		queryParams.Set("error_"+field, msg)
	}

	if len(queryParams) > 0 {
		return baseURL + "?" + queryParams.Encode()
	}
	return baseURL
}
