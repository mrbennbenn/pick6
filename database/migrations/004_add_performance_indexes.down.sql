-- Rollback performance indexes

DROP INDEX IF EXISTS idx_responses_session_id;
DROP INDEX IF EXISTS idx_slugs_event_id;
