
-- name: CreateMaterial :one
INSERT INTO material (
    course_id,
    title,
    material_file,
    order_number
) VALUES (
    $1, $2, $3, $4
) RETURNING *;


-- name: UpdateMaterial :one
UPDATE material
SET
    title = COALESCE(sqlc.narg(title),title),
    material_file = COALESCE(sqlc.narg(material_file),material_file)
WHERE
    material_id = sqlc.arg(material_id)
RETURNING *;


-- name: DeleteMaterial :exec
DELETE FROM material
WHERE material_id = $1;

-- name: GetTotalMaterialsInCourse :one
SELECT COUNT(*)
FROM material
WHERE course_id = $1;

-- name: GetMaterial :one
SELECT 
    m.*
FROM 
    material m
LEFT JOIN 
    courses c ON m.course_id = c.course_id
WHERE 
    c.course_id = $1
    AND m.material_id = $2;


-- name: ListMaterial :many
SELECT  
    material_id
    course_id,
    title,
    material_file,
    order_number
FROM material
WHERE course_id = $1
ORDER BY material_id
LIMIT 100;


-- name: ListMaterialByCourse :many
SELECT * FROM material
WHERE 
    course_id = $1
ORDER BY material_id
LIMIT 10;

-- name: GetMaterialByOrderNumber :one
SELECT * FROM material
WHERE 
    order_number = $1
    AND course_id = $2
LIMIT 1;    