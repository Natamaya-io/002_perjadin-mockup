DROP INDEX IF EXISTS idx_dalkot_record_officers_group;
DROP INDEX IF EXISTS uq_dalkot_record_officers_group;

ALTER TABLE dalkot_record_officers
    DROP CONSTRAINT IF EXISTS dalkot_record_officers_group_check;

ALTER TABLE dalkot_record_officers
    DROP COLUMN IF EXISTS officer_group;
