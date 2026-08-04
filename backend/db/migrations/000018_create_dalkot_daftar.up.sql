CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS dalkot_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id VARCHAR(64) NOT NULL UNIQUE,
    active_email VARCHAR(255) NOT NULL,
    service_date DATE NOT NULL,
    category VARCHAR(120) NOT NULL,
    accompanied_official VARCHAR(255) NOT NULL,
    dalkot_type VARCHAR(40) NOT NULL
        CHECK (dalkot_type IN ('SPJ_RILL', 'KEBIJAKAN_PROTOKOL')),
    policy_category VARCHAR(40)
        CHECK (
            policy_category IS NULL OR
            policy_category IN ('JAM_KERJA', 'OVERTIME', 'HARI_LIBUR')
        ),
    activity_name TEXT NOT NULL,
    location TEXT NOT NULL,
    report_content TEXT,
    status VARCHAR(30) NOT NULL DEFAULT 'DRAFT'
        CHECK (status IN ('DRAFT', 'DIAJUKAN', 'SELESAI', 'DIBATALKAN')),
    total_amount BIGINT NOT NULL DEFAULT 0 CHECK (total_amount >= 0),
    document_filename TEXT,
    document_original_name TEXT,
    document_mime VARCHAR(100),
    document_size BIGINT NOT NULL DEFAULT 0 CHECK (document_size >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS dalkot_record_officers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dalkot_id UUID NOT NULL
        REFERENCES dalkot_records(id)
        ON DELETE CASCADE,
    employee_id UUID,
    name VARCHAR(255) NOT NULL,
    nip VARCHAR(100),
    position VARCHAR(255),
    amount BIGINT NOT NULL DEFAULT 0 CHECK (amount >= 0),
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dalkot_records_service_date
    ON dalkot_records(service_date DESC);

CREATE INDEX IF NOT EXISTS idx_dalkot_records_type
    ON dalkot_records(dalkot_type);

CREATE INDEX IF NOT EXISTS idx_dalkot_records_status
    ON dalkot_records(status);

CREATE INDEX IF NOT EXISTS idx_dalkot_officers_dalkot_id
    ON dalkot_record_officers(dalkot_id);
