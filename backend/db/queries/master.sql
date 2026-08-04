-- name: GetProvinces :many
SELECT * FROM provinces WHERE deleted_at IS NULL ORDER BY name ASC;

-- name: GetProvinceByCode :one
SELECT * FROM provinces WHERE code = $1 AND deleted_at IS NULL LIMIT 1;

-- name: CreateProvince :one
INSERT INTO provinces (id, name, code) VALUES ($1, $2, $3) RETURNING *;

-- name: GetSBMRates :many
SELECT * FROM sbm_rates WHERE deleted_at IS NULL ORDER BY year DESC, province_id ASC;

-- name: GetSBMRateByProvinceAndYear :one
SELECT * FROM sbm_rates WHERE province_id = $1 AND year = $2 AND deleted_at IS NULL LIMIT 1;

-- name: CreateSBMRate :one
INSERT INTO sbm_rates (
  id, province_id, year, fullboard_rate, fullhalf_rate, outside_city_rate, inside_city_rate, diklat_rate, hotel_echelon1, hotel_echelon2, hotel_echelon3, hotel_echelon4, hotel_staff, taxi_rate
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
) RETURNING *;

-- name: GetSettingByKey :one
SELECT key, value, updated_at FROM settings WHERE key = $1 LIMIT 1;

-- name: GetSettings :many
SELECT key, value, updated_at FROM settings;

-- name: UpdateSetting :one
INSERT INTO settings (key, value, updated_at)
VALUES ($1, $2, CURRENT_TIMESTAMP)
ON CONFLICT (key) DO UPDATE SET
  value = EXCLUDED.value,
  updated_at = EXCLUDED.updated_at
RETURNING key, value, updated_at;
