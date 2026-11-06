-- name: InsertYoutubeIds :many
INSERT INTO youtube(id)
SELECT unnest(@ids::text[])
ON CONFLICT (id) DO NOTHING RETURNING id;

-- name: InsertTwitterIds :many
INSERT INTO twitter(id)
SELECT unnest(@ids::text[])
ON CONFLICT (id) DO NOTHING RETURNING id;

-- name: InsertMangaDexIds :many
INSERT INTO mangadex(id)
SELECT unnest(@ids::text[])
ON CONFLICT (id) DO NOTHING RETURNING id;