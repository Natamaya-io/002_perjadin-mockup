CREATE TABLE account_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(30) NOT NULL UNIQUE,
    mak VARCHAR(100) NOT NULL UNIQUE,
    description VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE procurement_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_code_id UUID NOT NULL
        REFERENCES account_codes(id) ON DELETE RESTRICT,
    name VARCHAR(180) NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE funding_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    year INTEGER NOT NULL,
    month_number INTEGER NOT NULL CHECK (
        month_number BETWEEN 1 AND 12
    ),
    month_name VARCHAR(20) NOT NULL,
    gup_label VARCHAR(20) NOT NULL,
    UNIQUE (year, month_number),
    UNIQUE (year, gup_label)
);

CREATE TABLE budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    year INTEGER NOT NULL,
    procurement_type_id UUID NOT NULL
        REFERENCES procurement_types(id) ON DELETE RESTRICT,
    amount NUMERIC(18,2) NOT NULL DEFAULT 0
        CHECK (amount >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (year, procurement_type_id)
);

CREATE TABLE monthly_ls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    funding_source_id UUID NOT NULL UNIQUE
        REFERENCES funding_sources(id) ON DELETE CASCADE,
    amount NUMERIC(18,2) NOT NULL DEFAULT 0
        CHECK (amount >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE gup_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id VARCHAR(50) NOT NULL UNIQUE,
    payment_description VARCHAR(255) NOT NULL,
    procurement_type_id UUID NOT NULL
        REFERENCES procurement_types(id) ON DELETE RESTRICT,
    funding_source_id UUID
        REFERENCES funding_sources(id) ON DELETE SET NULL,
    value_amount NUMERIC(18,2) NOT NULL DEFAULT 0,
    paid_amount NUMERIC(18,2) NOT NULL DEFAULT 0,
    tax_amount NUMERIC(18,2) NOT NULL DEFAULT 0,
    difference_amount NUMERIC(18,2)
        GENERATED ALWAYS AS (
            value_amount - paid_amount - tax_amount
        ) STORED,
    receipt_date DATE,
    recipient VARCHAR(200),
    pum VARCHAR(150),
    document_filename VARCHAR(255),
    document_original_name VARCHAR(255),
    document_mime VARCHAR(100),
    document_size BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE protocol_officers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    display_no INTEGER,
    name VARCHAR(180) NOT NULL UNIQUE,
    nip VARCHAR(50),
    position VARCHAR(255),
    grade VARCHAR(50),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE dalkot (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id VARCHAR(50) NOT NULL UNIQUE,
    execution_date DATE NOT NULL,
    category VARCHAR(50) NOT NULL,
    official VARCHAR(180) NOT NULL,
    dalkot_type VARCHAR(80) NOT NULL,
    activity_name VARCHAR(255) NOT NULL,
    location VARCHAR(255) NOT NULL,
    assignment_letter_no VARCHAR(120),
    assignment_letter_date DATE,
    report_content TEXT,
    total_spj_cost NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_actual_cost NUMERIC(18,2) NOT NULL DEFAULT 0,
    status VARCHAR(30) NOT NULL DEFAULT 'Draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (status IN ('Draft', 'Pending', 'Sukses'))
);

CREATE TABLE dalkot_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dalkot_id UUID NOT NULL
        REFERENCES dalkot(id) ON DELETE CASCADE,
    document_type VARCHAR(50) NOT NULL DEFAULT 'documentation',
    stored_filename VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255),
    mime_type VARCHAR(100),
    file_size BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE dalkot_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dalkot_id UUID NOT NULL
        REFERENCES dalkot(id) ON DELETE CASCADE,
    officer_id UUID
        REFERENCES protocol_officers(id) ON DELETE SET NULL,
    officer_name_snapshot VARCHAR(180) NOT NULL,
    assignment_type VARCHAR(10) NOT NULL,
    spj_cost NUMERIC(18,2) NOT NULL DEFAULT 0,
    actual_cost NUMERIC(18,2) NOT NULL DEFAULT 0,
    status VARCHAR(30) NOT NULL DEFAULT 'Pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (assignment_type IN ('SPJ', 'RIIL')),
    CHECK (status IN ('Draft', 'Pending', 'Sukses')),
    UNIQUE (
        dalkot_id,
        assignment_type,
        officer_name_snapshot
    )
);

CREATE INDEX idx_gup_receipt_date
    ON gup_transactions(receipt_date);

CREATE INDEX idx_dalkot_execution_date
    ON dalkot(execution_date);

CREATE INDEX idx_dalkot_assignment_officer
    ON dalkot_assignments(officer_id);