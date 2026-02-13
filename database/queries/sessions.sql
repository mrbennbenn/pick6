-- name: GetSession :one
SELECT * FROM sessions WHERE session_id = $1 LIMIT 1;

-- name: UpsertSession :one
INSERT INTO sessions (session_id, name, email, mobile)
VALUES ($1, $2, $3, $4)
ON CONFLICT (session_id) 
DO UPDATE SET 
    name = EXCLUDED.name,
    email = EXCLUDED.email,
    mobile = EXCLUDED.mobile
RETURNING *;
