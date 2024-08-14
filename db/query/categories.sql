-- name: CreateCategory :one
INSERT INTO categories (
    category
   
) VALUES (
    $1
)  RETURNING *;

-- name: GetCategory :one
SELECT * FROM categories
WHERE category_id = $1;

-- name: UpdateCategory :one
UPDATE categories
SET  category = $2
WHERE category_id = $1
RETURNING *;

-- name: DeleteCategory :exec
DELETE FROM categories
WHERE category_id = $1;


-- name: ListAllCategories :many
SELECT * FROM categories
WHERE category_id =$1
LIMIT $2
OFFSET $3;
