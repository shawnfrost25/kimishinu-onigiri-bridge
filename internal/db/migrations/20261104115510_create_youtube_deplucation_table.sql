-- Create "youtube" table
CREATE TABLE "public"."youtube" (
  "id" text NOT NULL,
  CONSTRAINT "youtube_id_pkey" PRIMARY KEY ("id"),
  CONSTRAINT "youtube_id_check" CHECK (length(id) = 11)
);
