-- NOTE: Forward-only migration strategy is enforced for production safety.
-- SAFE: rollback scripts are intentionally no-op to prevent accidental data loss.
SELECT 1;
