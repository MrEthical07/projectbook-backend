-- Reverses 000015. Drops the search-index triggers, functions and table plus the
-- global_feedback_submissions table. Triggers are dropped before the functions
-- they depend on. The search_index is a derived projection (rebuilt from the
-- artifact tables by the up), so dropping it loses nothing that cannot be
-- regenerated; global_feedback_submissions rows are removed with that table.

DROP TRIGGER IF EXISTS trg_search_index_stories ON stories;
DROP TRIGGER IF EXISTS trg_search_index_journeys ON journeys;
DROP TRIGGER IF EXISTS trg_search_index_problems ON problems;
DROP TRIGGER IF EXISTS trg_search_index_ideas ON ideas;
DROP TRIGGER IF EXISTS trg_search_index_tasks ON tasks;
DROP TRIGGER IF EXISTS trg_search_index_feedback ON feedback;
DROP TRIGGER IF EXISTS trg_search_index_resources ON resources;
DROP TRIGGER IF EXISTS trg_search_index_pages ON pages;
DROP TRIGGER IF EXISTS trg_search_index_calendar_events ON calendar_events;

DROP FUNCTION IF EXISTS search_index_sync_stories();
DROP FUNCTION IF EXISTS search_index_sync_journeys();
DROP FUNCTION IF EXISTS search_index_sync_problems();
DROP FUNCTION IF EXISTS search_index_sync_ideas();
DROP FUNCTION IF EXISTS search_index_sync_tasks();
DROP FUNCTION IF EXISTS search_index_sync_feedback();
DROP FUNCTION IF EXISTS search_index_sync_resources();
DROP FUNCTION IF EXISTS search_index_sync_pages();
DROP FUNCTION IF EXISTS search_index_sync_calendar_events();
DROP FUNCTION IF EXISTS search_index_upsert(uuid, text, uuid, text, text, text, text, timestamptz);
DROP FUNCTION IF EXISTS search_index_delete(uuid, text, uuid);
DROP FUNCTION IF EXISTS search_index_build_vector(text, text);

DROP TABLE IF EXISTS global_feedback_submissions;
DROP TABLE IF EXISTS search_index;
