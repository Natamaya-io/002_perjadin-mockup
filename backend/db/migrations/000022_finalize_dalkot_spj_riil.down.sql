DROP INDEX IF EXISTS idx_dalkot_record_officers_group;
DROP INDEX IF EXISTS uq_dalkot_record_officers_group;

ALTER TABLE dalkot_record_officers
    DROP CONSTRAINT IF EXISTS dalkot_record_officers_status_check;

ALTER TABLE dalkot_record_officers
    DROP CONSTRAINT IF EXISTS dalkot_record_officers_group_check;

UPDATE dalkot_record_officers
SET officer_group = 'RILL'
WHERE officer_group = 'RIIL';

ALTER TABLE dalkot_record_officers
    ADD CONSTRAINT dalkot_record_officers_group_check
    CHECK (officer_group IN ('SPJ', 'RILL'));

ALTER TABLE dalkot_records
    DROP CONSTRAINT IF EXISTS dalkot_records_status_check;

UPDATE dalkot_records
SET status = CASE
    WHEN status = 'Draft' THEN 'DRAFT'
    WHEN status = 'Pending' THEN 'DIAJUKAN'
    WHEN status = 'Sukses' THEN 'SELESAI'
    ELSE 'DRAFT'
END;

ALTER TABLE dalkot_records
    ADD CONSTRAINT dalkot_records_status_check
    CHECK (status IN ('DRAFT', 'DIAJUKAN', 'SELESAI', 'DIBATALKAN'));

ALTER TABLE dalkot_records
    DROP CONSTRAINT IF EXISTS dalkot_records_category_check;

UPDATE dalkot_records
SET total_amount = total_spj_amount + total_riil_amount;

ALTER TABLE dalkot_record_officers
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS grade;

ALTER TABLE dalkot_records
    DROP COLUMN IF EXISTS total_riil_amount,
    DROP COLUMN IF EXISTS total_spj_amount,
    DROP COLUMN IF EXISTS assignment_letter_date,
    DROP COLUMN IF EXISTS assignment_letter_number;
