CREATE TABLE IF NOT EXISTS dalkot_officer_master (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    nip VARCHAR(100),
    position VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_dalkot_officer_master_name_nip
    ON dalkot_officer_master (
        LOWER(name),
        COALESCE(LOWER(nip), '')
    );

INSERT INTO dalkot_officer_master (
    name,
    nip,
    position,
    sort_order
)
VALUES
    ('Nurin', '', 'Petugas Protokol', 1),
    ('Afen', '', 'Petugas Protokol', 2),
    ('Dhika', '', 'Petugas Protokol', 3)
ON CONFLICT DO NOTHING;
