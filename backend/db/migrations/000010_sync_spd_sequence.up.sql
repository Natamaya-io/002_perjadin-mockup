-- Sync the sequence again to account for records created while the old logic was active.
-- This ensures the sequence catches up to the true maximum ID-SPJ number currently in the table.
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
