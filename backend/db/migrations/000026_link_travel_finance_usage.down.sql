DROP TABLE IF EXISTS finance_usages;

DROP INDEX IF EXISTS idx_travel_records_finance_source;

ALTER TABLE travel_records
    DROP CONSTRAINT IF EXISTS fk_travel_records_finance_source;

ALTER TABLE travel_records
    DROP COLUMN IF EXISTS finance_source_id;
