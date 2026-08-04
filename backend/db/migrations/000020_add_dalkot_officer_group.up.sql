ALTER TABLE dalkot_record_officers
ADD COLUMN officer_group VARCHAR(10) NOT NULL DEFAULT 'SPJ';

ALTER TABLE dalkot_record_officers
ADD CONSTRAINT dalkot_record_officers_group_check
CHECK (officer_group IN ('SPJ', 'RILL'));

CREATE INDEX idx_dalkot_officers_group
ON dalkot_record_officers (dalkot_id, officer_group);