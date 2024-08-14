-- name: CreateProfilePicture :one
INSERT INTO profile_pictures(
    user_id,
    picture
) VALUES (
    $1, $2
) RETURNING *;

-- name: UpdateProfilePicture :one 
UPDATE profile_pictures
SET
    picture = COALESCE(sqlc.narg(picture),picture)
WHERE
    user_id = sqlc.arg(user_id)
RETURNING *;

-- name: DeleteProfilePicture :exec
DELETE FROM profile_pictures
WHERE user_id = $1;

-- name: GetProfilePicture :one
SELECT * FROM profile_pictures
WHERE user_id = $1;