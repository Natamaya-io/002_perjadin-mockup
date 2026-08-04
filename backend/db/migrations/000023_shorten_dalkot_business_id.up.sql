-- Ubah nomor bisnis Dalkot menjadi ringkas dan berurutan: DLK-001, DLK-002, dst.
-- Primary key UUID pada kolom `id` tetap dipertahankan dan tidak berubah.

-- Gunakan nilai sementara agar pembaruan massal tidak melanggar unique constraint.
UPDATE dalkot_records
SET business_id = 'TMP-' || id::text;

WITH numbered AS (
    SELECT
        id,
        ROW_NUMBER() OVER (ORDER BY created_at ASC, id ASC) AS sequence_number
    FROM dalkot_records
)
UPDATE dalkot_records AS records
SET business_id = 'DLK-' || LPAD(numbered.sequence_number::text, 3, '0')
FROM numbered
WHERE records.id = numbered.id;
