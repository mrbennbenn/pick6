-- Event-Level Engagement Queries

-- name: GetEventEngagementTotal :one
SELECT 
    COUNT(DISTINCT r.session_id) as sessions,
    COUNT(*) as total_votes
FROM responses r
JOIN questions q ON q.question_id = r.question_id
WHERE q.event_id = $1;

-- name: GetEventEngagementBySlug :many
SELECT 
    s.slug,
    COUNT(DISTINCT r.session_id) as sessions,
    COALESCE(COUNT(r.session_id), 0) as total_votes
FROM slugs s
LEFT JOIN responses r ON r.slug = s.slug
    AND r.question_id IN (SELECT question_id FROM questions WHERE questions.event_id = sqlc.arg(event_id))
WHERE s.event_id = sqlc.arg(event_id)
GROUP BY s.slug
ORDER BY s.slug;

-- name: GetEventRetentionBySlug :many
SELECT 
    s.slug,
    q.question_id,
    q.big_text,
    COUNT(DISTINCT r.session_id) as sessions_answered
FROM slugs s
CROSS JOIN questions q
LEFT JOIN responses r ON r.question_id = q.question_id AND r.slug = s.slug
WHERE s.event_id = $1 AND q.event_id = $1
GROUP BY s.slug, q.question_id, q.big_text
ORDER BY s.slug, q.question_id ASC;

-- Question-Level Engagement Queries

-- name: GetQuestionEngagementTotal :one
SELECT 
    COUNT(DISTINCT session_id) as sessions,
    COUNT(*) as total_votes,
    COALESCE(SUM(CASE WHEN choice = 'a' THEN 1 ELSE 0 END), 0) as votes_a,
    COALESCE(SUM(CASE WHEN choice = 'b' THEN 1 ELSE 0 END), 0) as votes_b
FROM responses
WHERE question_id = $1;

-- name: GetQuestionEngagementBySlug :many
SELECT 
    s.slug,
    COUNT(DISTINCT r.session_id) as sessions,
    COALESCE(COUNT(r.session_id), 0) as total_votes,
    COALESCE(SUM(CASE WHEN r.choice = 'a' THEN 1 ELSE 0 END), 0) as votes_a,
    COALESCE(SUM(CASE WHEN r.choice = 'b' THEN 1 ELSE 0 END), 0) as votes_b
FROM slugs s
LEFT JOIN responses r ON r.slug = s.slug AND r.question_id = $1
WHERE s.event_id = (SELECT event_id FROM questions WHERE question_id = $1)
GROUP BY s.slug
ORDER BY s.slug;
