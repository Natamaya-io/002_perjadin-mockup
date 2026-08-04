CREATE INDEX IF NOT EXISTS idx_travel_records_deleted_spd ON travel_records(deleted_at, spd_number);
CREATE INDEX IF NOT EXISTS idx_travel_records_deleted_at ON travel_records(deleted_at);
