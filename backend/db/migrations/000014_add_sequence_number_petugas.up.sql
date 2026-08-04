CREATE SEQUENCE IF NOT EXISTS travel_records_seq_no START WITH 471;
ALTER TABLE travel_records ADD COLUMN IF NOT EXISTS sequence_number INTEGER DEFAULT nextval('travel_records_seq_no');

WITH numbered AS (
  SELECT id, ROW_NUMBER() OVER (ORDER BY created_at ASC) as rn
  FROM travel_records
)
UPDATE travel_records tr
SET sequence_number = n.rn
FROM numbered n
WHERE tr.id = n.id AND (tr.sequence_number = 0 OR tr.sequence_number IS NULL OR tr.sequence_number > 1000);
