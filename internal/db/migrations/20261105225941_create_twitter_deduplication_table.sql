-- Create "twitter" table
CREATE TABLE "public"."twitter" (
  "id" text NOT NULL,
  CONSTRAINT "twitter_id_pkey" PRIMARY KEY ("id"),
  CONSTRAINT "twitter_id_check" CHECK (length(id) = 19)
);
