-- name: InsertYoutubeId :execrows
INSERT INTO youtube(id)
VALUES (
    sqlc.arg(id)::text
) ON CONFLICT (id) DO NOTHING;