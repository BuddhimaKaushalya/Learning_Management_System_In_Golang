-- name: CreateCourseProgress :one
INSERT INTO course_progress (
    course_id,
    user_id,
    progress
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetCourseProgress :one
SELECT * FROM course_progress
WHERE 
    user_id = $1
    AND course_id = $2
LIMIT 1;


-- name: ListCourseProgressByUser :many
SELECT * FROM course_progress
WHERE user_id =$1
ORDER BY courseprogress_id
LIMIT $2
OFFSET $3;

-- name: DeleteCourseProgress :exec
DELETE FROM course_progress
WHERE courseprogress_id = $1;

-- name: UpdateCourseProgress :one
UPDATE course_progress
SET 
    course_id = COALESCE(sqlc.narg(course_id), course_id),
    progress = COALESCE(sqlc.narg(progress),progress)
WHERE
    user_id = sqlc.arg(user_id)
RETURNING *;

-- name: GetCourseCompletedUserCount :one
SELECT COUNT(*) 
FROM course_progress
WHERE progress = $1;

-- name: GetInProgressCourseCount :one
SELECT COUNT(*) 
FROM course_progress
WHERE progress > 0
    AND progress < 100 ;

-- name: GetTotalUserCompletedCourseCount :one
SELECT COUNT(*) 
FROM course_progress
WHERE 
    progress = $1
    AND user_id = $2;

