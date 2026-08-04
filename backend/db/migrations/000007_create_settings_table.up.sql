CREATE TABLE IF NOT EXISTS settings (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO settings (key, value) VALUES 
('ppk_name', 'Arief Hafidiyanto'),
('ppk_nip', '19720827 200312 1 002'),
('bendahara_name', 'Liana Setyawati'),
('bendahara_nip', '19800512 200901 2 001')
ON CONFLICT (key) DO NOTHING;
