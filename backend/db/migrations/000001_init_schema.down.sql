ALTER TABLE travel_reports DROP CONSTRAINT IF EXISTS travel_reports_record_id_key;
DROP TABLE IF EXISTS travel_reports;

ALTER TABLE travel_costs DROP CONSTRAINT IF EXISTS travel_costs_record_id_key;
DROP TABLE IF EXISTS travel_costs;

DROP TABLE IF EXISTS travel_records;
DROP TABLE IF EXISTS sbm_rates;
DROP TABLE IF EXISTS provinces;
DROP TABLE IF EXISTS users;

-- Optional: DROP EXTENSION IF EXISTS "pgcrypto";
