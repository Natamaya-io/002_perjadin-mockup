CREATE TABLE finance_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    year INTEGER NOT NULL CHECK (year BETWEEN 2000 AND 2100),
    code VARCHAR(50) NOT NULL,
    name VARCHAR(180) NOT NULL,
    amount NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (amount >= 0),
    notes TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (year, code)
);

CREATE TABLE finance_allocations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    finance_source_id UUID NOT NULL
        REFERENCES finance_sources(id) ON DELETE RESTRICT,
    allocation_type VARCHAR(10) NOT NULL CHECK (allocation_type IN ('GUP', 'LS')),
    amount NUMERIC(18,2) NOT NULL CHECK (amount > 0),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_finance_sources_year
    ON finance_sources(year);

CREATE INDEX idx_finance_allocations_source
    ON finance_allocations(finance_source_id);

CREATE INDEX idx_finance_allocations_type
    ON finance_allocations(allocation_type);
