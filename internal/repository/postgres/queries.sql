-- name: GetByShortKey :one
SELECT long_url
FROM url_shortener.links 
WHERE short_key = sqlc.arg(short_key);

-- name: SaveShortKey :execresult
INSERT INTO url_shortener.links (short_key, long_url)
VALUES (sqlc.arg(short_key), sqlc.arg(long_url))
ON CONFLICT (short_key) DO NOTHING;


