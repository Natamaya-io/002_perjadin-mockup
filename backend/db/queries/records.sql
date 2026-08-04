-- name: CreateTravelRecord :one
INSERT INTO travel_records (
  id, spd_number, employee_id, creator_id, start_date, end_date, location, province, type, purpose, stakeholder, agenda, status, is_viewed, report_status, payment_status, total_cost, surat_tugas_path, surat_tugas_number, surat_tugas_date
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
) RETURNING *;

-- name: GetTravelRecords :many
SELECT * FROM travel_records
WHERE deleted_at IS NULL
  AND (NULLIF($1::text, '') IS NULL OR status = $1)
ORDER BY created_at DESC
LIMIT 1000;

-- name: GetTravelRecordByID :one
SELECT * FROM travel_records WHERE id = $1 AND deleted_at IS NULL LIMIT 1;

-- name: NextSpdNumber :one
-- Returns the next unique sequence value for SPD number generation.
-- nextval() is atomic and safe under concurrent load.
SELECT nextval('spd_number_seq')::BIGINT AS next_val;

-- name: GetOverlappingRecords :many
SELECT * FROM travel_records 
WHERE employee_id = $1 
  AND start_date <= sqlc.arg('new_end_date')
  AND end_date >= sqlc.arg('new_start_date')
  AND status != 'Rejected'
  AND deleted_at IS NULL;

-- name: UpdateTravelRecord :one
UPDATE travel_records SET
  spd_number = $2,
  employee_id = $3,
  creator_id = $4,
  start_date = $5,
  end_date = $6,
  location = $7,
  province = $8,
  type = $9,
  purpose = $10,
  stakeholder = $11,
  agenda = $12,
  status = $13,
  is_viewed = $14,
  report_status = $15,
  payment_status = $16,
  total_cost = $17,
  surat_tugas_path = $18,
  surat_tugas_number = $19,
  surat_tugas_date = $20,
  updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteTravelRecord :exec
DELETE FROM travel_records WHERE id = $1;

-- name: DeleteTravelRecordBySpd :exec
DELETE FROM travel_records WHERE spd_number = $1;

-- name: DeleteTravelLocationsBySpd :exec
DELETE FROM travel_locations
WHERE travel_record_id IN (SELECT id FROM travel_records WHERE spd_number = $1);

-- name: CreateTravelCost :one
INSERT INTO travel_costs (
  travel_record_id, ticket_go, ticket_back, daily_allowance_days, daily_allowance_rate, hotel_days, hotel_rate, local_transport, regional_transport, transport_mode, transport_amount, other_cost, other_cost_desc, receipt_files, ticket_go_file, ticket_back_file, boarding_pass_file, hotel_file, transport_file, additional_costs, details
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21
) RETURNING *;

-- name: GetTravelCostByRecordID :one
SELECT * FROM travel_costs WHERE travel_record_id = $1 LIMIT 1;

-- name: UpdateTravelCost :one
UPDATE travel_costs SET
  ticket_go = $2, ticket_back = $3, daily_allowance_days = $4, daily_allowance_rate = $5, hotel_days = $6, hotel_rate = $7, local_transport = $8, regional_transport = $9, transport_mode = $10, transport_amount = $11, other_cost = $12, other_cost_desc = $13, receipt_files = $14, ticket_go_file = $15, ticket_back_file = $16, boarding_pass_file = $17, hotel_file = $18, transport_file = $19, additional_costs = $20, details = $21
WHERE travel_record_id = $1
RETURNING *;

-- name: DeleteTravelCost :exec
DELETE FROM travel_costs WHERE travel_record_id = $1;

-- name: CreateTravelReport :one
INSERT INTO travel_reports (
  travel_record_id, text, submitted_at, files, sppd_file, surat_tugas_file, ppk_name, ppk_nip, bendahara_name, bendahara_nip, tanggal_merah
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetTravelReportByRecordID :one
SELECT * FROM travel_reports WHERE travel_record_id = $1 LIMIT 1;

-- name: UpdateTravelReport :one
UPDATE travel_reports SET
  text = $2, submitted_at = $3, files = $4, sppd_file = $5, surat_tugas_file = $6, ppk_name = $7, ppk_nip = $8, bendahara_name = $9, bendahara_nip = $10, tanggal_merah = $11
WHERE travel_record_id = $1
RETURNING *;

-- name: DeleteTravelReport :exec
DELETE FROM travel_reports WHERE travel_record_id = $1;

-- name: CreateTravelLocation :one
INSERT INTO travel_locations (
  id, travel_record_id, location, province, start_date, end_date
) VALUES (
  $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetTravelLocationsByRecordID :many
SELECT * FROM travel_locations WHERE travel_record_id = $1 AND deleted_at IS NULL ORDER BY start_date ASC, created_at ASC;

-- name: DeleteTravelLocationsByRecordID :exec
DELETE FROM travel_locations WHERE travel_record_id = $1;

-- name: SyncSpdSequence :exec
DO $$
DECLARE
    max_id bigint;
    has_records boolean;
BEGIN
    SELECT MAX(CAST(NULLIF(regexp_replace(spd_number, '\D', '', 'g'), '') AS BIGINT)) INTO max_id FROM travel_records WHERE spd_number ILIKE 'ID-SPJ-%' AND deleted_at IS NULL;
    SELECT COUNT(*) > 0 INTO has_records FROM travel_records WHERE deleted_at IS NULL;
    
    IF has_records THEN
        PERFORM setval('spd_number_seq', COALESCE(max_id, 1), true);
    ELSE
        PERFORM setval('spd_number_seq', 1, false);
    END IF;
END $$;

-- name: GetDashboardStatusCounts :many
SELECT status, payment_status, report_status, COUNT(*) as count 
FROM travel_records 
WHERE deleted_at IS NULL
  AND (NULLIF(sqlc.narg('user_id')::uuid, NULL) IS NULL OR employee_id = sqlc.narg('user_id') OR creator_id = sqlc.narg('user_id'))
GROUP BY status, payment_status, report_status;

-- name: GetDashboardReportCounts :many
SELECT report_status, COUNT(*) as count 
FROM travel_records 
WHERE deleted_at IS NULL
  AND (NULLIF(sqlc.narg('user_id')::uuid, NULL) IS NULL OR employee_id = sqlc.narg('user_id') OR creator_id = sqlc.narg('user_id'))
GROUP BY report_status;

-- name: GetActiveTripsCount :one
SELECT COUNT(*) 
FROM travel_records 
WHERE deleted_at IS NULL 
  AND status IN ('Approved', 'Submitted', 'Assigned', 'Draft')
  AND start_date <= CURRENT_TIMESTAMP 
  AND end_date >= CURRENT_TIMESTAMP
  AND (NULLIF(sqlc.narg('user_id')::uuid, NULL) IS NULL OR employee_id = sqlc.narg('user_id') OR creator_id = sqlc.narg('user_id'));

-- name: GetTotalTripsCount :one
SELECT COUNT(DISTINCT spd_number) 
FROM travel_records 
WHERE deleted_at IS NULL
  AND (NULLIF(sqlc.narg('user_id')::uuid, NULL) IS NULL OR employee_id = sqlc.narg('user_id') OR creator_id = sqlc.narg('user_id'));

-- name: GetRecentRecords :many
SELECT * FROM travel_records 
WHERE deleted_at IS NULL 
  AND (NULLIF(sqlc.narg('user_id')::uuid, NULL) IS NULL OR employee_id = sqlc.narg('user_id') OR creator_id = sqlc.narg('user_id'))
ORDER BY created_at DESC, id ASC 
LIMIT 5;

-- name: GetDashboardBudgets :many
SELECT 
  EXTRACT(YEAR FROM COALESCE(start_date, created_at))::INT as year,
  EXTRACT(MONTH FROM COALESCE(start_date, created_at))::INT as month,
  SUM(total_cost)::FLOAT8 as total
FROM travel_records
WHERE deleted_at IS NULL
  AND (NULLIF(sqlc.narg('user_id')::uuid, NULL) IS NULL OR employee_id = sqlc.narg('user_id') OR creator_id = sqlc.narg('user_id'))
GROUP BY year, month
ORDER BY year DESC, month DESC;

-- name: GetPaginatedSPDs :many
SELECT travel_records.spd_number
FROM travel_records
LEFT JOIN users ON travel_records.employee_id = users.id
WHERE travel_records.deleted_at IS NULL
  AND (NULLIF(sqlc.narg('status')::text, '') IS NULL OR travel_records.status = sqlc.narg('status'))
  AND (NULLIF(sqlc.narg('report_status')::text, '') IS NULL OR travel_records.report_status = sqlc.narg('report_status'))
  AND (NULLIF(sqlc.narg('payment_status')::text, '') IS NULL OR travel_records.payment_status = sqlc.narg('payment_status'))
  AND (
    NULLIF(sqlc.narg('search')::text, '') IS NULL
    OR travel_records.spd_number ILIKE '%' || sqlc.narg('search') || '%'
    OR travel_records.location ILIKE '%' || sqlc.narg('search') || '%'
    OR users.name ILIKE '%' || sqlc.narg('search') || '%'
  )
  AND (NULLIF(sqlc.narg('start_date')::timestamp, NULL) IS NULL OR travel_records.start_date >= sqlc.narg('start_date'))
  AND (NULLIF(sqlc.narg('end_date')::timestamp, NULL) IS NULL OR travel_records.start_date <= sqlc.narg('end_date'))
  AND (NULLIF(sqlc.narg('user_id')::uuid, NULL) IS NULL OR travel_records.employee_id = sqlc.narg('user_id') OR travel_records.creator_id = sqlc.narg('user_id'))
  AND (NULLIF(sqlc.narg('cursor')::text, '') IS NULL OR travel_records.spd_number < sqlc.narg('cursor'))
GROUP BY travel_records.spd_number
ORDER BY travel_records.spd_number DESC
LIMIT sqlc.arg('limit');

-- name: GetTotalPaginatedSPDsCount :one
SELECT COUNT(DISTINCT travel_records.spd_number)
FROM travel_records
LEFT JOIN users ON travel_records.employee_id = users.id
WHERE travel_records.deleted_at IS NULL
  AND (NULLIF(sqlc.narg('status')::text, '') IS NULL OR travel_records.status = sqlc.narg('status'))
  AND (NULLIF(sqlc.narg('report_status')::text, '') IS NULL OR travel_records.report_status = sqlc.narg('report_status'))
  AND (NULLIF(sqlc.narg('payment_status')::text, '') IS NULL OR travel_records.payment_status = sqlc.narg('payment_status'))
  AND (
    NULLIF(sqlc.narg('search')::text, '') IS NULL
    OR travel_records.spd_number ILIKE '%' || sqlc.narg('search') || '%'
    OR travel_records.location ILIKE '%' || sqlc.narg('search') || '%'
    OR users.name ILIKE '%' || sqlc.narg('search') || '%'
  )
  AND (NULLIF(sqlc.narg('start_date')::timestamp, NULL) IS NULL OR travel_records.start_date >= sqlc.narg('start_date'))
  AND (NULLIF(sqlc.narg('end_date')::timestamp, NULL) IS NULL OR travel_records.start_date <= sqlc.narg('end_date'))
  AND (NULLIF(sqlc.narg('user_id')::uuid, NULL) IS NULL OR travel_records.employee_id = sqlc.narg('user_id') OR travel_records.creator_id = sqlc.narg('user_id'));
-- name: GetRecordsBySPDs :many
SELECT * FROM travel_records
WHERE deleted_at IS NULL
  AND spd_number = ANY(sqlc.arg('spds')::text[])
ORDER BY created_at DESC, id ASC;

-- name: GetTravelCostsByRecordIDs :many
SELECT * FROM travel_costs 
WHERE travel_record_id = ANY(sqlc.arg('record_ids')::uuid[]);

-- name: GetTravelReportsByRecordIDs :many
SELECT * FROM travel_reports 
WHERE travel_record_id = ANY(sqlc.arg('record_ids')::uuid[]);

-- name: GetTravelLocationsByRecordIDs :many
SELECT * FROM travel_locations 
WHERE travel_record_id = ANY(sqlc.arg('record_ids')::uuid[]) 
  AND deleted_at IS NULL 
;

-- name: GetSequenceNumbers :many
WITH ordered AS (
    SELECT id, ROW_NUMBER() OVER (
        ORDER BY CAST(SUBSTRING(spd_number FROM '[0-9]+') AS INTEGER) ASC, sequence_number ASC, created_at ASC
    ) as actual_rank 
    FROM travel_records 
    WHERE deleted_at IS NULL AND spd_number IS NOT NULL AND spd_number <> ''
)
SELECT id, actual_rank::int AS sequence_number 
FROM ordered 
WHERE id = ANY(sqlc.arg('record_ids')::uuid[]);
