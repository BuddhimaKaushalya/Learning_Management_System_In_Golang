-- name: CreateCourses :one
INSERT INTO courses (
    user_id,
    title,
    description,
    image,
    catagory,
    what_will,
    sequential_access
    
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetCourses :one
SELECT * FROM courses
WHERE course_id = $1;

-- name: GetCourseByUserID :one
SELECT * FROM courses
WHERE user_id = $1;


-- name: UpdateCourses :one
UPDATE courses
SET  
       title = COALESCE(sqlc.narg(title), title),
       image = COALESCE(sqlc.narg(image), image),
       description = COALESCE(sqlc.narg(description), description),
       catagory = COALESCE(sqlc.narg(catagory), catagory),
        sequential_access = COALESCE(sqlc.narg(sequential_access), sequential_access),
         what_will = COALESCE(sqlc.narg(what_will), what_will)

WHERE  course_id = sqlc.arg(course_id)
RETURNING *;

-- name: ListCourses :many
SELECT * FROM courses
ORDER BY course_id
LIMIT $1
OFFSET $2;

-- name: DeleteCourses :exec
DELETE FROM courses
WHERE course_id = $1 ;

-- name: GetTotalCourseCount :one
SELECT COUNT(*) 
FROM courses;

-- name: GetEntireCourse :one
SELECT 
    c.*,
    COALESCE(json_agg(DISTINCT m) FILTER (WHERE m.material_id IS NOT NULL), '[]') AS material,
    COALESCE(json_agg(DISTINCT a) FILTER (WHERE a.assignment_id IS NOT NULL), '[]') AS assignment
FROM 
    courses c
LEFT JOIN 
    material m ON c.course_id = m.course_id
LEFT JOIN 
    assignment a ON c.course_id = a.course_id
WHERE 
    c.course_id = $1
GROUP BY 
    c.course_id
LIMIT 1;

-- name: ListAllCourseByCatagory :many
SELECT  
    course_id,
    user_id,
    title,
    description,
    image,
    catagory,
    created_at,
    what_will,
    updated_at
FROM courses
WHERE catagory = $1
ORDER BY course_id
LIMIT 100;

-- name: ListAllCourseCatagories :many
SELECT DISTINCT  
    catagory
FROM courses
LIMIT 20;






