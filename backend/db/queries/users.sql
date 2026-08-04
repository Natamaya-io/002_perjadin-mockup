-- name: CreateUser :one
INSERT INTO users (
  id, email, password, name, role, nip, nomor_hp, pangkat, golongan, jabatan, tingkat_biaya, session_id, demo_password
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL LIMIT 1;

-- name: UpdateUser :one
UPDATE users SET
  email = $2,
  password = $3,
  name = $4,
  role = $5,
  nip = $6,
  nomor_hp = $7,
  pangkat = $8,
  golongan = $9,
  jabatan = $10,
  tingkat_biaya = $11,
  session_id = $12,
  demo_password = $13,
  updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: GetUsers :many
SELECT * FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC;

-- name: DeleteUser :exec
UPDATE users SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1;

-- name: GetUsersByIDs :many
SELECT * FROM users WHERE id = ANY(sqlc.arg('ids')::uuid[]) AND deleted_at IS NULL;
