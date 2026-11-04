CREATE TABLE youtube (
    id TEXT NOT NULL,
    CONSTRAINT youtube_id_check CHECK (length(id) = 11),
    CONSTRAINT youtube_id_pkey PRIMARY KEY (id)
);