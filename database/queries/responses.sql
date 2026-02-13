-- name: UpsertResponse :one
INSERT INTO responses (question_id, session_id, slug, choice, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
ON CONFLICT (question_id, session_id)
DO UPDATE SET
    choice = EXCLUDED.choice,
    slug = EXCLUDED.slug,
    updated_at = NOW()
RETURNING *;

-- name: GetResponsesBySessionAndEvent :many
SELECT r.question_id, r.session_id, r.slug, r.choice, r.created_at, r.updated_at
FROM responses r
JOIN questions q ON q.question_id = r.question_id
WHERE r.session_id = $1 AND q.event_id = $2;
