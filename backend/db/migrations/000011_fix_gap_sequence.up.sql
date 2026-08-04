-- Hard delete the ghost/soft-deleted records >= 240 that were generated during testing
-- This removes the "dirty" numbers from the database.
DELETE FROM travel_records 
WHERE deleted_at IS NOT NULL 
AND CAST(NULLIF(regexp_replace(spd_number, '\D', '', 'g'), '') AS BIGINT) >= 240;

-- Renumber any currently active records that got a skipped number (e.g., 245 -> 240)
WITH recent AS (
    SELECT id, ROW_NUMBER() OVER (ORDER BY created_at ASC) - 1 as rn
    FROM travel_records
    WHERE deleted_at IS NULL AND CAST(NULLIF(regexp_replace(spd_number, '\D', '', 'g'), '') AS BIGINT) >= 240
)
UPDATE travel_records
SET spd_number = 'ID-SPJ-' || LPAD((240 + recent.rn)::text, 3, '0')
FROM recent
WHERE travel_records.id = recent.id;

-- Resynchronize the sequence to the new strict maximum
DO $$
DECLARE
    max_id bigint;
    has_records boolean;
BEGIN
    SELECT MAX(CAST(NULLIF(regexp_replace(spd_number, '\D', '', 'g'), '') AS BIGINT)) INTO max_id FROM travel_records WHERE spd_number ILIKE 'ID-SPJ-%';
    SELECT COUNT(*) > 0 INTO has_records FROM travel_records;
    
    IF has_records THEN
        PERFORM setval('spd_number_seq', COALESCE(max_id, 1), true);
    ELSE
        PERFORM setval('spd_number_seq', 1, false);
    END IF;
END $$;
