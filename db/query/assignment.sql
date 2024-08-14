-- name: CreateAssignment :one
INSERT INTO assignment (
    title,
    course_id,
    due_date,
    assignment_file
) VALUES (
    $1, $2, $3, $4
)  RETURNING *;

-- name: GetAssignment :one
SELECT * FROM assignment
WHERE assignment_id = $1;

-- name: UpdateAssignment :one
UPDATE assignment
SET  title = $2,due_date = $3, assignment_file = $4, course_id = $5
WHERE assignment_id = $1
RETURNING *;

-- name: DeleteAssignment :exec
DELETE FROM assignment
WHERE assignment_id = $1;