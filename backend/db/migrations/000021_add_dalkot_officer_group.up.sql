ALTER TABLE dalkot_record_officers
    ADD COLUMN IF NOT EXISTS officer_group VARCHAR(10) NOT NULL DEFAULT 'SPJ';

ALTER TABLE dalkot_record_officers
    DROP CONSTRAINT IF EXISTS dalkot_record_officers_group_check;

ALTER TABLE dalkot_record_officers
    ADD CONSTRAINT dalkot_record_officers_group_check
    CHECK (officer_group IN ('SPJ', 'RILL'));

DO $$
DECLARE
    item RECORD;
BEGIN
    FOR item IN
        SELECT c.conname
        FROM pg_constraint c
        WHERE c.conrelid = 'dalkot_record_officers'::regclass
          AND c.contype = 'u'
          AND pg_get_constraintdef(c.oid) ILIKE '%(dalkot_id, employee_id)%'
          AND pg_get_constraintdef(c.oid) NOT ILIKE '%officer_group%'
    LOOP
        EXECUTE format(
            'ALTER TABLE dalkot_record_officers DROP CONSTRAINT %I',
            item.conname
        );
    END LOOP;
END $$;

DROP INDEX IF EXISTS uq_dalkot_record_officers_group;

CREATE UNIQUE INDEX uq_dalkot_record_officers_group
    ON dalkot_record_officers (
        dalkot_id,
        employee_id,
        officer_group
    )
    WHERE employee_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_dalkot_record_officers_group
    ON dalkot_record_officers (
        dalkot_id,
        officer_group
    );
