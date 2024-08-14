-- name: CreateRequest :one
INSERT INTO requests (
    course_id,
    confirm,
    pending

) VALUES (
    $1, $2, $3
) RETURNING *;


-- name: UpdateRequest :one
UPDATE requests
SET 
    confirm = COALESCE(sqlc.narg(confirm), confirm),
    pending = COALESCE(sqlc.narg(pending), pending)
WHERE
    course_id = sqlc.arg(course_id)
RETURNING *;

-- name: GetRequest :one
SELECT * FROM requests
WHERE request_id = $1 LIMIT 1;