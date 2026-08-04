-- Finalisasi struktur Dalkot agar mengikuti alur gup-mvc:
-- biaya SPJ dan Riil terpisah, assignment SPJ/RIIL, status per petugas,
-- dan data Surat Tugas.

ALTER TABLE dalkot_records
    ADD COLUMN IF NOT EXISTS assignment_letter_number VARCHAR(120),
    ADD COLUMN IF NOT EXISTS assignment_letter_date DATE,
    ADD COLUMN IF NOT EXISTS total_spj_amount BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_riil_amount BIGINT NOT NULL DEFAULT 0;

ALTER TABLE dalkot_record_officers
    ADD COLUMN IF NOT EXISTS grade VARCHAR(100),
    ADD COLUMN IF NOT EXISTS status VARCHAR(30) NOT NULL DEFAULT 'Pending';

-- Samakan kategori lama dengan kategori Dalkot pada gup-mvc.
UPDATE dalkot_records
SET category = CASE
    WHEN policy_category = 'JAM_KERJA' THEN 'Jam Kerja'
    WHEN policy_category = 'OVERTIME' THEN 'Overtime'
    WHEN policy_category = 'HARI_LIBUR' THEN 'Hari Libur'
    WHEN category = 'JAM_KERJA' THEN 'Jam Kerja'
    WHEN category = 'OVERTIME' THEN 'Overtime'
    WHEN category = 'HARI_LIBUR' THEN 'Hari Libur'
    WHEN category IN ('Jam Kerja', 'Overtime', 'Hari Libur') THEN category
    ELSE 'Jam Kerja'
END;

ALTER TABLE dalkot_records
    DROP CONSTRAINT IF EXISTS dalkot_records_category_check;

ALTER TABLE dalkot_records
    ADD CONSTRAINT dalkot_records_category_check
    CHECK (category IN ('Jam Kerja', 'Overtime', 'Hari Libur'));

-- Samakan status pengajuan dengan gup-mvc.
ALTER TABLE dalkot_records
    DROP CONSTRAINT IF EXISTS dalkot_records_status_check;

UPDATE dalkot_records
SET status = CASE
    WHEN status IN ('Draft', 'DRAFT') THEN 'Draft'
    WHEN status IN ('Pending', 'PENDING', 'DIAJUKAN') THEN 'Pending'
    WHEN status IN ('Sukses', 'SUKSES', 'SELESAI') THEN 'Sukses'
    WHEN status = 'DIBATALKAN' THEN 'Draft'
    ELSE 'Draft'
END;

ALTER TABLE dalkot_records
    ADD CONSTRAINT dalkot_records_status_check
    CHECK (status IN ('Draft', 'Pending', 'Sukses'));

-- Gunakan kode RIIL seperti pada gup-mvc, bukan RILL.
ALTER TABLE dalkot_record_officers
    DROP CONSTRAINT IF EXISTS dalkot_record_officers_group_check;

UPDATE dalkot_record_officers
SET officer_group = 'RIIL'
WHERE officer_group = 'RILL';

ALTER TABLE dalkot_record_officers
    ADD CONSTRAINT dalkot_record_officers_group_check
    CHECK (officer_group IN ('SPJ', 'RIIL'));

ALTER TABLE dalkot_record_officers
    DROP CONSTRAINT IF EXISTS dalkot_record_officers_status_check;

UPDATE dalkot_record_officers
SET status = CASE
    WHEN status IN ('Draft', 'DRAFT') THEN 'Draft'
    WHEN status IN ('Pending', 'PENDING', 'DIAJUKAN') THEN 'Pending'
    WHEN status IN ('Sukses', 'SUKSES', 'SELESAI') THEN 'Sukses'
    ELSE 'Pending'
END;

ALTER TABLE dalkot_record_officers
    ADD CONSTRAINT dalkot_record_officers_status_check
    CHECK (status IN ('Draft', 'Pending', 'Sukses'));

-- User yang sama boleh menjadi SPJ dan RIIL pada pengajuan yang sama,
-- tetapi tidak boleh diduplikasi dalam kelompok yang sama.
DROP INDEX IF EXISTS uq_dalkot_record_officers_group;

CREATE UNIQUE INDEX uq_dalkot_record_officers_group
    ON dalkot_record_officers (
        dalkot_id,
        employee_id,
        officer_group
    )
    WHERE employee_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_dalkot_record_officers_group
    ON dalkot_record_officers (dalkot_id, officer_group);

-- Hitung ulang total terpisah dari rincian petugas yang sudah ada.
UPDATE dalkot_records AS dr
SET
    total_spj_amount = COALESCE((
        SELECT SUM(dro.amount)
        FROM dalkot_record_officers AS dro
        WHERE dro.dalkot_id = dr.id
          AND dro.officer_group = 'SPJ'
    ), 0),
    total_riil_amount = COALESCE((
        SELECT SUM(dro.amount)
        FROM dalkot_record_officers AS dro
        WHERE dro.dalkot_id = dr.id
          AND dro.officer_group = 'RIIL'
    ), 0),
    total_amount = 0;
