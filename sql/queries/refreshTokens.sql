-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (
    token,
    user_id,
    expires_at
) VALUES (
    $1, $2, $3
);

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW(),
    updated_at = NOW()
WHERE token = $1;

-- name: GetRefreshToken :one
SELECT *
FROM refresh_tokens
WHERE token = $1;
