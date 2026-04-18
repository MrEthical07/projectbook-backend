DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'journey_status') THEN
        ALTER TYPE journey_status ADD VALUE IF NOT EXISTS 'Locked';
    END IF;
END
$$;