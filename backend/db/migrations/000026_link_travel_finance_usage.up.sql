ALTER TABLE travel_records
    ADD COLUMN IF NOT EXISTS finance_source_id UUID;

ALTER TABLE travel_records
    DROP CONSTRAINT IF EXISTS fk_travel_records_finance_source;

ALTER TABLE travel_records
    ADD CONSTRAINT fk_travel_records_finance_source
    FOREIGN KEY (finance_source_id)
    REFERENCES finance_sources(id)
    ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_travel_records_finance_source
    ON travel_records(finance_source_id);

CREATE TABLE IF NOT EXISTS finance_usages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    spd_number VARCHAR(100) NOT NULL UNIQUE,
    finance_source_id UUID NOT NULL
        REFERENCES finance_sources(id)
        ON DELETE RESTRICT,
    budget_type VARCHAR(10) NOT NULL
        CHECK (budget_type IN ('GUP', 'LS')),
    amount NUMERIC(18,2) NOT NULL DEFAULT 0
        CHECK (amount >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_finance_usages_source
    ON finance_usages(finance_source_id);

CREATE INDEX IF NOT EXISTS idx_finance_usages_budget_type
    ON finance_usages(budget_type);

CREATE INDEX IF NOT EXISTS idx_finance_usages_source_type
    ON finance_usages(finance_source_id, budget_type);
