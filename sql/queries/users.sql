-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: UpdateAccount :exec
UPDATE users
SET email = $1,
    hashed_password = $2,
    updated_at = NOW()
WHERE id = $3;

-- name: ActivateChirpyRed :exec
UPDATE users
SET is_chirpy_red = TRUE,
    updated_at = NOW()
WHERE id = $1;

-- name: DeactivateChirpyRed :exec
UPDATE users
SET is_chirpy_red = FALSE,
    updated_at = NOW()
WHERE id = $1;
