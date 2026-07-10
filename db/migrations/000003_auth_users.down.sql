-- Reverses 000003. Drops the users table and its indexes.
-- Safe only after every table that references users has already been dropped by
-- the downs of later migrations (000005, 000006, 000013, 000015, ...), which is
-- the case when rolling the whole chain back. Running this in isolation while
-- dependents still exist will (correctly) fail rather than silently cascade.
DROP TABLE IF EXISTS users;
