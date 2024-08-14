-- name: CreateUser :one
INSERT INTO users (
    user_name,
    first_name,
    last_name,
    hashed_password,
    email,
    role,
    is_email_verified
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;


-- name: GetUser :one
SELECT * FROM users
WHERE user_name = $1 LIMIT 1;


-- name: UpdateUser :one
UPDATE users
SET 
    hashed_password = COALESCE(sqlc.narg(hashed_password), hashed_password),
    password_changed_at = COALESCE(sqlc.narg(password_changed_at), password_changed_at),
    first_name = COALESCE(sqlc.narg(first_name), first_name),
    last_name = COALESCE(sqlc.narg(last_name), last_name),
    email = COALESCE(sqlc.narg(email), email),
    is_email_verified = COALESCE(sqlc.narg(is_email_verified), is_email_verified),
    user_name = COALESCE(sqlc.narg(user_name), user_name),
    role = COALESCE(sqlc.narg(role), role)
WHERE
    user_id = sqlc.arg(user_id)
RETURNING *;

-- name: ListUser :many
SELECT * FROM users
WHERE role = $1
ORDER BY user_id
LIMIT $2
OFFSET $3;

-- name: StudentCount :one
SELECT COUNT(DISTINCT user_id) AS Count
FROM users
WHERE role = $1;


-- name: DeleteUsers :exec
DELETE FROM users
WHERE user_id = $1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE user_id = $1 LIMIT 1;


-- name: GetTotalUserCount :one
SELECT COUNT(*) FROM users
WHERE role = $1;

-- name: GetTotalRegisteredUserCount :one
SELECT COUNT(*) FROM users
WHERE is_email_verified = $1;


-- name: UpdateUsersPassword :one
UPDATE users
SET
    hashed_password = COALESCE(sqlc.narg(hashed_password), hashed_password),
    password_changed_at = COALESCE(sqlc.narg(password_changed_at), password_changed_at),
    updated_at = COALESCE(sqlc.narg(updated_at), updated_at)
WHERE 
    email = sqlc.arg(email)
RETURNING *;

-- name: CheckEmail :one
SELECT u.email
FROM users u
WHERE EXISTS (
    SELECT *
    FROM users
    WHERE u.email = $1
);
