-- Add performance indexes to optimize queries under load
-- These indexes significantly reduce query time for session-based lookups

-- Index on responses.session_id speeds up GetResponsesBySessionAndEvent query
-- which is called on every question page view
CREATE INDEX IF NOT EXISTS idx_responses_session_id ON responses(session_id);

-- Index on slugs.event_id speeds up GetEventBySlug JOIN query
-- which is called on every authenticated request
CREATE INDEX IF NOT EXISTS idx_slugs_event_id ON slugs(event_id);
