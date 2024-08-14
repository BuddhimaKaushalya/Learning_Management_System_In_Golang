-- name: CreateUserStatus :one
INSERT INTO user_status (
    user_id,
    active,
    pending
   
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetUserStatus :one
SELECT * FROM user_status
WHERE user_id = $1;

-- name: ListUserStatus :many
SELECT * FROM user_status
WHERE user_id = $1
ORDER BY status_id
LIMIT $2
OFFSET $3;

-- name: DeleteUserStatus :exec
DELETE FROM user_status
WHERE status_id = $1;


-- name: UpdateUserStatus :one
UPDATE user_status
SET 
    status_id = COALESCE(sqlc.narg(status_id), status_id),
    active = COALESCE(sqlc.narg(active), active),
    pending = TRUE
WHERE
    user_id = sqlc.arg(user_id)
RETURNING *;

-- name: UpdateUserStatusByAdmin :one
UPDATE user_status
SET 
    pending = $1,
    active = $2
WHERE
    status_id = sqlc.arg(status_id)  
RETURNING *;