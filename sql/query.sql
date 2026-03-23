-- name: DatabasePing :one
SELECT 'pong'::text AS message;

-- name: ListRegions :many
SELECT region_id, region_name
FROM regions
ORDER BY region_id
LIMIT $1;

-- name: GetRegion :one
SELECT region_id, region_name
FROM regions
WHERE region_id = $1;

-- name: CreateRegion :one
INSERT INTO regions (region_id, region_name)
VALUES ($1, $2)
RETURNING region_id, region_name;

-- name: UpdateRegion :one
UPDATE regions
SET region_name = $2
WHERE region_id = $1
RETURNING region_id, region_name;

-- name: DeleteRegion :one
DELETE FROM regions
WHERE region_id = $1
RETURNING region_id, region_name;
