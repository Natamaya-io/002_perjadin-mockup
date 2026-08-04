-- Data master GUP untuk tahun 2026.
-- Migration ini aman dijalankan berulang karena memakai ON CONFLICT.

INSERT INTO account_codes (code, mak)
VALUES
    ('S.521119', '2158.01.WA.2158.EBA.994.002.S.521119'),
    ('S.523121', '2158.01.WA.2158.EBA.994.002.S.523121'),
    ('S.522141', '2158.01.WA.2158.EBA.994.002.S.522141'),
    ('S.524111', '2158.01.WA.2158.EBA.994.002.S.524111'),
    ('S.524113', '2158.01.WA.2158.EBA.994.002.S.524113'),
    ('S.524211', '2158.01.WA.2158.EBA.994.002.S.524211')
ON CONFLICT (code) DO UPDATE
SET mak = EXCLUDED.mak;

INSERT INTO procurement_types (account_code_id, name, is_active)
SELECT id, 'VIP Halim', TRUE
FROM account_codes
WHERE code = 'S.521119'
ON CONFLICT (name) DO UPDATE
SET account_code_id = EXCLUDED.account_code_id,
    is_active = TRUE;

INSERT INTO procurement_types (account_code_id, name, is_active)
SELECT id, 'Pass Bandara', TRUE
FROM account_codes
WHERE code = 'S.521119'
ON CONFLICT (name) DO UPDATE
SET account_code_id = EXCLUDED.account_code_id,
    is_active = TRUE;

INSERT INTO procurement_types (account_code_id, name, is_active)
SELECT id, 'Langganan AI', TRUE
FROM account_codes
WHERE code = 'S.521119'
ON CONFLICT (name) DO UPDATE
SET account_code_id = EXCLUDED.account_code_id,
    is_active = TRUE;

INSERT INTO procurement_types (account_code_id, name, is_active)
SELECT id, 'Pemeliharaan PC', TRUE
FROM account_codes
WHERE code = 'S.523121'
ON CONFLICT (name) DO UPDATE
SET account_code_id = EXCLUDED.account_code_id,
    is_active = TRUE;

INSERT INTO procurement_types (account_code_id, name, is_active)
SELECT id, 'Pemeliharaan Printer', TRUE
FROM account_codes
WHERE code = 'S.523121'
ON CONFLICT (name) DO UPDATE
SET account_code_id = EXCLUDED.account_code_id,
    is_active = TRUE;

INSERT INTO procurement_types (account_code_id, name, is_active)
SELECT id, 'Pemeliharaan Notebook', TRUE
FROM account_codes
WHERE code = 'S.523121'
ON CONFLICT (name) DO UPDATE
SET account_code_id = EXCLUDED.account_code_id,
    is_active = TRUE;

INSERT INTO procurement_types (account_code_id, name, is_active)
SELECT id, 'Pemeliharaan Mobil Operasional Protokol', TRUE
FROM account_codes
WHERE code = 'S.523121'
ON CONFLICT (name) DO UPDATE
SET account_code_id = EXCLUDED.account_code_id,
    is_active = TRUE;

INSERT INTO procurement_types (account_code_id, name, is_active)
SELECT id, 'Sewa Mobil Menaker', TRUE
FROM account_codes
WHERE code = 'S.522141'
ON CONFLICT (name) DO UPDATE
SET account_code_id = EXCLUDED.account_code_id,
    is_active = TRUE;

INSERT INTO procurement_types (account_code_id, name, is_active)
SELECT id, 'Sewa Kendaraan Protokol', TRUE
FROM account_codes
WHERE code = 'S.522141'
ON CONFLICT (name) DO UPDATE
SET account_code_id = EXCLUDED.account_code_id,
    is_active = TRUE;

INSERT INTO procurement_types (account_code_id, name, is_active)
SELECT id, 'Perjalanan Dinas Dalam Negeri', TRUE
FROM account_codes
WHERE code = 'S.524111'
ON CONFLICT (name) DO UPDATE
SET account_code_id = EXCLUDED.account_code_id,
    is_active = TRUE;

INSERT INTO procurement_types (account_code_id, name, is_active)
SELECT id, 'Perjalanan Dinas Dalam Kota', TRUE
FROM account_codes
WHERE code = 'S.524113'
ON CONFLICT (name) DO UPDATE
SET account_code_id = EXCLUDED.account_code_id,
    is_active = TRUE;

INSERT INTO procurement_types (account_code_id, name, is_active)
SELECT id, 'Luar Negeri', TRUE
FROM account_codes
WHERE code = 'S.524211'
ON CONFLICT (name) DO UPDATE
SET account_code_id = EXCLUDED.account_code_id,
    is_active = TRUE;

INSERT INTO funding_sources (year, month_number, month_name, gup_label)
VALUES
    (2026, 1, 'Januari', 'GUP 1'),
    (2026, 2, 'Februari', 'GUP 2'),
    (2026, 3, 'Maret', 'GUP 3'),
    (2026, 4, 'April', 'GUP 4'),
    (2026, 5, 'Mei', 'GUP 5'),
    (2026, 6, 'Juni', 'GUP 6'),
    (2026, 7, 'Juli', 'GUP 7'),
    (2026, 8, 'Agustus', 'GUP 8'),
    (2026, 9, 'September', 'GUP 9'),
    (2026, 10, 'Oktober', 'GUP 10'),
    (2026, 11, 'November', 'GUP 11'),
    (2026, 12, 'Desember', 'GUP 12')
ON CONFLICT (year, month_number) DO UPDATE
SET month_name = EXCLUDED.month_name,
    gup_label = EXCLUDED.gup_label;

-- Anggaran awal mengikuti aplikasi PHP lama.
INSERT INTO budgets (year, procurement_type_id, amount)
SELECT 2026, id,
    CASE name
        WHEN 'VIP Halim' THEN 6000000
        WHEN 'Pass Bandara' THEN 4000000
        WHEN 'Langganan AI' THEN 5000000
        WHEN 'Pemeliharaan PC' THEN 10000000
        WHEN 'Pemeliharaan Printer' THEN 6000000
        WHEN 'Pemeliharaan Notebook' THEN 9000000
        WHEN 'Pemeliharaan Mobil Operasional Protokol' THEN 13000000
        WHEN 'Sewa Mobil Menaker' THEN 20000000
        WHEN 'Sewa Kendaraan Protokol' THEN 16000000
        WHEN 'Perjalanan Dinas Dalam Negeri' THEN 15000000
        WHEN 'Perjalanan Dinas Dalam Kota' THEN 5000000
        WHEN 'Luar Negeri' THEN 25000000
        ELSE 0
    END
FROM procurement_types
WHERE name IN (
    'VIP Halim',
    'Pass Bandara',
    'Langganan AI',
    'Pemeliharaan PC',
    'Pemeliharaan Printer',
    'Pemeliharaan Notebook',
    'Pemeliharaan Mobil Operasional Protokol',
    'Sewa Mobil Menaker',
    'Sewa Kendaraan Protokol',
    'Perjalanan Dinas Dalam Negeri',
    'Perjalanan Dinas Dalam Kota',
    'Luar Negeri'
)
ON CONFLICT (year, procurement_type_id) DO UPDATE
SET amount = EXCLUDED.amount,
    updated_at = CURRENT_TIMESTAMP;

INSERT INTO monthly_ls (funding_source_id, amount)
SELECT id,
    CASE WHEN month_number = 6 THEN 100000000 ELSE 0 END
FROM funding_sources
WHERE year = 2026
ON CONFLICT (funding_source_id) DO NOTHING;