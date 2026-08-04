CREATE INDEX IF NOT EXISTS idx_travel_records_spd_numeric ON travel_records( (CAST(SUBSTRING(spd_number FROM '[0-9]+') AS INTEGER)) );
