ALTER TABLE travel_records
    ADD COLUMN IF NOT EXISTS budget_source VARCHAR(10);

ALTER TABLE travel_records
    DROP CONSTRAINT IF EXISTS chk_travel_records_budget_source;

ALTER TABLE travel_records
    ADD CONSTRAINT chk_travel_records_budget_source
    CHECK (budget_source IS NULL OR budget_source IN ('GUP', 'LS'));

COMMENT ON COLUMN travel_records.budget_source IS
    'Sumber anggaran perjalanan dinas yang dipilih saat review rincian biaya: GUP atau LS.';
