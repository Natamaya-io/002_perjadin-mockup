ALTER TABLE travel_records DROP COLUMN IF EXISTS sequence_number;
DROP SEQUENCE IF EXISTS travel_records_seq_no;
