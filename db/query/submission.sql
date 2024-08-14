-- name: CreateSubmission :one
INSERT INTO submission (
    assignment_id,
    user_id,
    grade,
    resource,
    date_of_submission,
    submitted
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetsubmissionsByAssignment :one
SELECT * FROM submission
WHERE assignment_id = $1 
LIMIT 1;

-- name: GetsubmissionsByUser :one
SELECT * FROM submission
WHERE user_id = $1
LIMIT 1;

-- name: Listsubmissions :many
SELECT * FROM submission
ORDER BY submission_id
LIMIT $1
OFFSET $2;

-- name: UpdateSubmission :one
UPDATE submission
SET assignment_id = $2 AND user_id = $3
WHERE submission_id = $1
RETURNING *;

-- name: DeleteSubmission :exec
DELETE FROM submission
WHERE 
    assignment_id = $1
    AND user_id = $2;


-- name: GetSubmission :one
SELECT * FROM submission
WHERE assignment_id = $1 AND user_id = $2;