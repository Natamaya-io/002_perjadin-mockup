CREATE INDEX idx_travel_records_status ON travel_records(status);
CREATE INDEX idx_travel_records_report_status ON travel_records(report_status);
CREATE INDEX idx_travel_records_payment_status ON travel_records(payment_status);
CREATE INDEX idx_travel_records_start_date ON travel_records(start_date);
CREATE INDEX idx_travel_records_end_date ON travel_records(end_date);
CREATE INDEX idx_travel_records_employee_id ON travel_records(employee_id);
CREATE INDEX idx_travel_records_creator_id ON travel_records(creator_id);
CREATE INDEX idx_travel_records_created_at ON travel_records(created_at);
