-- Reverses 000013. Drops task_assignees, the archived_from_status columns, the
-- feedback.status column and the feedback_status type. The task_assignees backfill
-- the up performed is discarded with the table (expected). feedback_status is
-- dropped only after both columns that use it (feedback.status and
-- feedback.archived_from_status) are removed.
DROP TABLE IF EXISTS task_assignees;

ALTER TABLE stories DROP COLUMN IF EXISTS archived_from_status;
ALTER TABLE journeys DROP COLUMN IF EXISTS archived_from_status;
ALTER TABLE problems DROP COLUMN IF EXISTS archived_from_status;
ALTER TABLE ideas DROP COLUMN IF EXISTS archived_from_status;
ALTER TABLE pages DROP COLUMN IF EXISTS archived_from_status;
ALTER TABLE resources DROP COLUMN IF EXISTS archived_from_status;
ALTER TABLE feedback DROP COLUMN IF EXISTS archived_from_status;

ALTER TABLE feedback DROP COLUMN IF EXISTS status;

DROP TYPE IF EXISTS feedback_status;
