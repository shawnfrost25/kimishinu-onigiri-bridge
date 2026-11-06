-- Create "mangadex" table
CREATE TABLE "public"."mangadex" (
  "id" text NOT NULL,
  CONSTRAINT "mangadex_id_pkey" PRIMARY KEY ("id"),
  CONSTRAINT "mangadex_id_check" CHECK (length(id) = 36)
);
