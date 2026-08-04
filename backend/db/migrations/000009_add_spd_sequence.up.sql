-- Create a sequence for atomic, race-free SPD number generation.
-- Seed it from the current maximum to prevent collisions with existing records.
-- The COALESCE default of 0 handles an empty table gracefully (next value = 1).
CREATE SEQUENCE IF NOT EXISTS spd_number_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

-- Advance the sequence past the highest number already in use so the first
-- nextval() call returns current_max + 1.
DO $$
DECLARE
    max_id bigint;
    has_records boolean;
BEGIN
    SELECT MAX(CAST(SUBSTRING(spd_number FROM 8) AS BIGINT)) INTO max_id FROM travel_records WHERE spd_number ~ '^ID-SPJ-[0-9]+$';
    SELECT COUNT(*) > 0 INTO has_records FROM travel_records;
    
    IF has_records THEN
        PERFORM setval('spd_number_seq', COALESCE(max_id, 1), true);
    ELSE
        PERFORM setval('spd_number_seq', 1, false);
    END IF;
END $$;
