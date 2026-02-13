-- name: GetEventBySlug :one
SELECT e.event_id, e.description, e.created_at
FROM events e
JOIN slugs s ON s.event_id = e.event_id
WHERE s.slug = $1;

-- name: GetEventByID :one
SELECT event_id, description, created_at
FROM events
WHERE event_id = $1;

-- name: GetQuestionByID :one
SELECT question_id, event_id, big_text, small_text, image_filename, choice_a, choice_b
FROM questions
WHERE question_id = $1;

-- name: ListQuestionsByEventID :many
SELECT question_id, event_id, big_text, small_text, image_filename, choice_a, choice_b
FROM questions
WHERE event_id = $1
ORDER BY question_id ASC;

-- name: GetQuestionByEventAndIndex :one
SELECT question_id, event_id, big_text, small_text, image_filename, choice_a, choice_b
FROM questions
WHERE event_id = sqlc.arg(event_id)
ORDER BY question_id ASC
LIMIT 1 OFFSET sqlc.arg(question_index) - 1;
