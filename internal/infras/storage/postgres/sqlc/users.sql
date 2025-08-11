-- name: CreateUser :exec
INSERT INTO users (id, first_name, last_name, email, password, role, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetUserByID :one
SELECT id, first_name, last_name, email, password, role, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, first_name, last_name, email, password, role, created_at, updated_at
FROM users
WHERE email = $1;

-- name: ListUsers :many
SELECT id, first_name, last_name, email, password, role, created_at, updated_at
FROM users
ORDER BY created_at DESC;

-- name: UpdateUser :exec
UPDATE users
SET first_name = $2,
    last_name  = $3,
    email      = $4,
    password   = $5,
    role       = $6,
    updated_at = $7
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;


