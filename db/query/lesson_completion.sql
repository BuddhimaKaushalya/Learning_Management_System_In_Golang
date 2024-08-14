-- name: CreateLessonCompletion :one
INSERT INTO lesson_completion (
    course_id,
    user_id,
    material_id,
    completed
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: UpdateLessonCompletion :one
UPDATE lesson_completion
SET  
       completed = COALESCE(sqlc.narg(completed), completed)
WHERE  user_id = sqlc.arg(user_id) AND course_id = sqlc.arg(course_id) AND material_id = sqlc.arg(material_id)
RETURNING *;


-- name: GetCompletedLessonsCount :one
SELECT COUNT(DISTINCT material_id) AS completed_lessons_count
FROM lesson_completion
WHERE user_id = $1 AND course_id = $2 AND completed = true;

-- name: DeleteLessonCompletion :exec
DELETE FROM lesson_completion
WHERE completion_id = $1;

-- name: GetLessonCompletion :one
SELECT * FROM lesson_completion
WHERE 
    user_id = $1
    AND course_id = $2
    AND material_id = $3
LIMIT 1;




