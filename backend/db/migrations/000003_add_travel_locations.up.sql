CREATE TABLE travel_locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    travel_record_id UUID NOT NULL REFERENCES travel_records(id) ON DELETE CASCADE,
    location VARCHAR(255) NOT NULL,
    province VARCHAR(255) NOT NULL,
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE NOT NULL
);
CREATE INDEX idx_travel_locations_deleted_at ON travel_locations(deleted_at);
CREATE INDEX idx_travel_locations_record_id ON travel_locations(travel_record_id);

-- Migrate existing data
INSERT INTO travel_locations (travel_record_id, location, province, start_date, end_date)
SELECT id, location, province, start_date, end_date
FROM travel_records
WHERE location IS NOT NULL AND start_date IS NOT NULL AND end_date IS NOT NULL;
