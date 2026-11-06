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

-- name: InsertMangaDexId :execrows
INSERT INTO mangadex(id)
VALUES (
    sqlc.arg(id)::text
) ON CONFLICT (id) DO NOTHING;