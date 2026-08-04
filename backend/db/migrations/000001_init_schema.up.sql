CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'protokol',
    nip VARCHAR(100),
    nomor_hp VARCHAR(50),
    pangkat VARCHAR(100),
    golongan VARCHAR(50),
    jabatan VARCHAR(255),
    tingkat_biaya VARCHAR(50),
    session_id VARCHAR(255),
    demo_password VARCHAR(255)
);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

CREATE TABLE provinces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL UNIQUE,
    code VARCHAR(50) UNIQUE
);
CREATE INDEX idx_provinces_deleted_at ON provinces(deleted_at);

CREATE TABLE sbm_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    province_id UUID NOT NULL REFERENCES provinces(id) ON DELETE RESTRICT,
    year INT NOT NULL,
    fullboard_rate DOUBLE PRECISION,
    fullhalf_rate DOUBLE PRECISION,
    outside_city_rate DOUBLE PRECISION,
    inside_city_rate DOUBLE PRECISION,
    diklat_rate DOUBLE PRECISION,
    hotel_echelon1 DOUBLE PRECISION,
    hotel_echelon2 DOUBLE PRECISION,
    hotel_echelon3 DOUBLE PRECISION,
    hotel_echelon4 DOUBLE PRECISION,
    hotel_staff DOUBLE PRECISION,
    taxi_rate DOUBLE PRECISION
);
CREATE INDEX idx_sbm_rates_deleted_at ON sbm_rates(deleted_at);
CREATE INDEX idx_sbm_rates_province_id ON sbm_rates(province_id);
CREATE INDEX idx_sbm_rates_year ON sbm_rates(year);

CREATE TABLE travel_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    spd_number VARCHAR(100),
    employee_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    location VARCHAR(255),
    province VARCHAR(255),
    type VARCHAR(50),
    purpose TEXT,
    stakeholder VARCHAR(255),
    agenda TEXT,
    status VARCHAR(50) DEFAULT 'Draft',
    is_viewed BOOLEAN DEFAULT false,
    report_status VARCHAR(50) DEFAULT 'Pending',
    payment_status VARCHAR(50) DEFAULT 'Unpaid',
    total_cost DOUBLE PRECISION,
    surat_tugas_path VARCHAR(255),
    surat_tugas_number VARCHAR(100)
);
CREATE INDEX idx_travel_records_deleted_at ON travel_records(deleted_at);
CREATE INDEX idx_travel_records_spd_number ON travel_records(spd_number);

CREATE TABLE travel_costs (
    travel_record_id UUID PRIMARY KEY REFERENCES travel_records(id) ON DELETE CASCADE,
    ticket_go DOUBLE PRECISION,
    ticket_back DOUBLE PRECISION,
    daily_allowance_days INT,
    daily_allowance_rate DOUBLE PRECISION,
    hotel_days INT,
    hotel_rate DOUBLE PRECISION,
    local_transport DOUBLE PRECISION,
    regional_transport DOUBLE PRECISION,
    transport_mode VARCHAR(100),
    transport_amount DOUBLE PRECISION,
    other_cost DOUBLE PRECISION,
    other_cost_desc TEXT,
    receipt_files JSONB,
    ticket_go_file JSONB,
    ticket_back_file JSONB,
    boarding_pass_file JSONB,
    hotel_file JSONB,
    transport_file JSONB,
    additional_costs JSONB
);
ALTER TABLE travel_costs ADD CONSTRAINT travel_costs_record_id_key UNIQUE (travel_record_id);

CREATE TABLE travel_reports (
    travel_record_id UUID PRIMARY KEY REFERENCES travel_records(id) ON DELETE CASCADE,
    text TEXT,
    submitted_at TIMESTAMP WITH TIME ZONE,
    files JSONB,
    sppd_file JSONB,
    surat_tugas_file JSONB
);
ALTER TABLE travel_reports ADD CONSTRAINT travel_reports_record_id_key UNIQUE (travel_record_id);
