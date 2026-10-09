ALTER TABLE travel_records
    DROP CONSTRAINT IF EXISTS chk_travel_records_budget_source;

ALTER TABLE travel_records
    DROP COLUMN IF EXISTS budget_source;
