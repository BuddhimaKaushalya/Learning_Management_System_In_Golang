
-- name: CreateMark :one
INSERT INTO marks (
    course_id,
    user_id,
    marks
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetMark :one
SELECT * FROM marks
WHERE mark_id = $1;


-- name: UpdateMark :one
UPDATE marks
SET marks = $2 , user_id = $3, course_id = $4
WHERE mark_id = $1
RETURNING *;

-- name: ListMarks :many
SELECT * FROM marks
WHERE course_id = $1
ORDER BY user_id
LIMIT $2
OFFSET $3;

-- name: DeleteMark :exec
DELETE FROM marks
WHERE mark_id = $1;
