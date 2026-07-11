-- Reverses 000001. Structural-only rollback: drops the system_settings table.
-- Rows written after this migration are removed with the table (inherent to
-- reverting a table creation); no state that predated the migration is lost.
DROP TABLE IF EXISTS system_settings;
