-- name: InsertYoutubeId :execrows
INSERT INTO youtube(id)
VALUES (
    sqlc.arg(id)::text
) ON CONFLICT (id) DO NOTHING;

-- name: InsertTwitterId :execrows
INSERT INTO twitter(id)
VALUES (
    sqlc.arg(id)::text
) ON CONFLICT (id) DO NOTHING;