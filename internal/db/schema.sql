CREATE TABLE youtube (
    id TEXT NOT NULL,
    CONSTRAINT youtube_id_check CHECK (length(id) = 11),
    CONSTRAINT youtube_id_pkey PRIMARY KEY (id)
);

CREATE TABLE twitter (
    id TEXT NOT NULL,
    CONSTRAINT twitter_id_check CHECK (length(id) = 19),
    CONSTRAINT twitter_id_pkey PRIMARY KEY (id)
);