-- name: CreateSubscription :one
INSERT INTO subscriptions (
    user_id,
    course_id,
    active,
    pending
) VALUES (
    $1, $2, $3, $4
) RETURNING *;



-- name: UpdateSubscriptions :one
UPDATE subscriptions
SET 
    pending = COALESCE(sqlc.narg(pending), pending),
    active = COALESCE(sqlc.narg(active), active)

WHERE
    user_id = sqlc.arg(user_id) AND course_id = sqlc.arg(course_id)
RETURNING *;



-- name: ListSubscriptionsByUser :many
SELECT * FROM subscriptions
WHERE course_id = $1
ORDER BY user_id
LIMIT $2
OFFSET $3;

-- name: ListSubscriptionsByCourse :many
SELECT * FROM subscriptions
WHERE user_id = $1
ORDER BY course_id
LIMIT $2
OFFSET $3;

-- name: GetSubscription :one
SELECT * FROM subscriptions
WHERE user_id = $1 ;


-- name: GetStudentCountInCourse :many
SELECT COUNT(DISTINCT user_id) AS Count
FROM courses
GROUP BY course_id
ORDER BY course_id;

-- name: GetTotalSubscribedUserCount :one
SELECT COUNT(*) 
FROM Subscriptions
WHERE active = $1;

-- name: GetUserCountForCertianCourse :one
SELECT COUNT(DISTINCT user_id) 
FROM subscriptions
WHERE 
    active = $1
    AND course_id = $2;

-- name: GetSubscriptionByUser :one
SELECT 
    s.subscription_id,
    s.user_id,
    s.course_id,
    u.user_name,
    u.first_name,
    u.last_name,
    s.active,
    s.pending,
    cp.progress,
    c.title
FROM subscriptions s
LEFT JOIN users u ON u.user_id = s.user_id
LEFT JOIN course_progress cp ON cp.user_id = s.user_id
LEFT JOIN courses c ON c.course_id = s.course_id
WHERE s.user_id = $1 AND s.course_id = $2
LIMIT 1;